// Package main fournit un serveur Git Smart HTTP minimal basé sur go-git,
// prenant en charge le pull (upload-pack) et le push (receive-pack) sans dépendance externe.
// Package main fournit un serveur Git Smart HTTP minimal basé sur go-git et Fiber v2,
// prenant en charge le pull (upload-pack) et le push (receive-pack).
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	gliderssh "github.com/gliderlabs/ssh"
	billyos "github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/gofiber/fiber/v2"
	gossh "golang.org/x/crypto/ssh"
)

// BaseDir est le répertoire où stocker les dépôts bare (*.git)
var BaseDir = "./repositories"

// StorageBackend selects where repositories are stored.
// Supported: "fs" (filesystem). "s3" is reserved (not implemented yet).
var StorageBackend = "fs"

func main() {
	if err := os.MkdirAll(BaseDir, 0755); err != nil {
		log.Fatal("création du répertoire base: ", err)
	}

	app := NewGitHTTPApp()
	go func() {
		log.Println("Serveur Git HTTP (go-git + Fiber) sur http://localhost:8080")
		if err := app.Listen(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	// Démarre le serveur SSH Git
	log.Println("Serveur Git SSH (go-git) sur ssh://localhost:2222")
	if err := startSSHServer(":2222"); err != nil {
		log.Fatal(err)
	}
}

// NewGitHTTPApp construit une application Fiber servant les endpoints Smart HTTP.
func NewGitHTTPApp() *fiber.App {
	app := fiber.New()
	// API: création de dépôt
	app.Post("/api/repos", createRepoHandler)
	// Route catch-all pour router selon les suffixes
	app.All("/*", gitFiberHandler)
	return app
}

// gitFiberHandler route les endpoints Smart HTTP Git vers les handlers adaptés.
func gitFiberHandler(c *fiber.Ctx) error {
	path := string(c.Request().URI().Path())
	method := string(c.Request().Header.Method())

	switch {
	case strings.HasSuffix(path, "/info/refs") && method == fiber.MethodGet:
		return handleInfoRefs(c)
	case strings.HasSuffix(path, "/git-upload-pack") && method == fiber.MethodPost:
		return handleUploadPack(c)
	case strings.HasSuffix(path, "/git-receive-pack") && method == fiber.MethodPost:
		return handleReceivePack(c)
	default:
		return c.SendString("ogit: serveur Git minimal (pull/push) via go-git")
	}
}

// ----- Handlers -----

// handleInfoRefs gère l'annonce des références (advertised-refs) pour git-upload-pack et git-receive-pack.
func handleInfoRefs(c *fiber.Ctx) error {
	service := c.Query("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		return c.Status(fiber.StatusBadRequest).SendString("paramètre service invalide")
	}

	repoPath := extractRepoDirFromPath(string(c.Request().URI().Path()), "/info/refs")
	if repoPath == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// Ne crée plus automatiquement les dépôts
	if !pathExists(repoPath) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	srv, ep, err := repoServer(repoPath)
	if err != nil || srv == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("erreur serveur dépôt: " + err.Error())
	}

	c.Set("Cache-Control", "no-cache")
	switch service {
	case "git-upload-pack":
		c.Set("Content-Type", "application/x-git-upload-pack-advertisement")
		sess, err := srv.NewUploadPackSession(ep, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		adv, err := sess.AdvertisedReferences()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := writeServiceAdvertisement(c.Response().BodyWriter(), service); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := adv.Encode(c.Response().BodyWriter()); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	case "git-receive-pack":
		c.Set("Content-Type", "application/x-git-receive-pack-advertisement")
		sess, err := srv.NewReceivePackSession(ep, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		adv, err := sess.AdvertisedReferences()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := writeServiceAdvertisement(c.Response().BodyWriter(), service); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := adv.Encode(c.Response().BodyWriter()); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	}
	return nil
}

// handleUploadPack traite les requêtes POST /git-upload-pack (clone/fetch).
func handleUploadPack(c *fiber.Ctx) error {
	repoPath := extractRepoDirFromPath(string(c.Request().URI().Path()), "/git-upload-pack")
	if repoPath == "" || !pathExists(repoPath) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	srv, ep, err := repoServer(repoPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	sess, err := srv.NewUploadPackSession(ep, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	req := packp.NewUploadPackRequest()
	if err := req.Decode(bytes.NewReader(c.Body())); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	resp, err := sess.UploadPack(context.Background(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer resp.Close()

	c.Set("Content-Type", "application/x-git-upload-pack-result")
	if err := resp.Encode(c.Response().BodyWriter()); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return nil
}

// handleReceivePack traite les requêtes POST /git-receive-pack (push).
func handleReceivePack(c *fiber.Ctx) error {
	repoPath := extractRepoDirFromPath(string(c.Request().URI().Path()), "/git-receive-pack")
	if repoPath == "" || !pathExists(repoPath) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	srv, ep, err := repoServer(repoPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	sess, err := srv.NewReceivePackSession(ep, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(bytes.NewReader(c.Body())); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	report, err := sess.ReceivePack(context.Background(), req)
	c.Set("Content-Type", "application/x-git-receive-pack-result")
	if err != nil {
		_ = report.Encode(c.Response().BodyWriter())
		return nil
	}
	if err := report.Encode(c.Response().BodyWriter()); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return nil
}

// ----- Helpers -----

// repoServer prépare un storer + serveur go-git pour le répertoire bare donné
// repoServer construit un transport serveur go-git + endpoint pour un dépôt bare.
func repoServer(repoPath string) (transport.Transport, *transport.Endpoint, error) {
	st, err := getStorerForPath(repoPath)
	if err != nil {
		return nil, nil, err
	}

	// Loader qui retourne toujours ce storer
	loader := &constLoader{st: st}
	srv := server.NewServer(loader)
	ep := &transport.Endpoint{Path: "/" + filepath.Base(repoPath)}
	return srv, ep, nil
}

// getStorerForPath retourne un storer selon le backend sélectionné.
func getStorerForPath(repoPath string) (storer.Storer, error) {
	switch StorageBackend {
	case "fs":
		fs := billyos.New(repoPath)
		return filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), nil
	case "s3":
		// TODO: implémenter un storer S3 (non disponible par défaut)
		return nil, fiber.NewError(fiber.StatusNotImplemented, "backend s3 non implémenté")
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "backend de stockage inconnu")
	}
}

// constLoader implémente server.Loader en retournant un storer fixe
// constLoader implémente server.Loader en retournant toujours le même storer.
type constLoader struct{ st storer.Storer }

func (c *constLoader) Load(ep *transport.Endpoint) (storer.Storer, error) {
	return c.st, nil
}

// Helpers
// extractRepoDirFromPath convertit une URL Smart HTTP en chemin de dépôt bare.
// Exemple: /test.git/info/refs -> ./repositories/test.git
func extractRepoDirFromPath(urlPath, suffix string) string {
	// enlève le suffixe ("/info/refs", etc.)
	idx := strings.LastIndex(urlPath, suffix)
	if idx < 0 {
		return ""
	}
	base := strings.TrimSuffix(urlPath[:idx], "/")
	rel := strings.TrimPrefix(base, "/")
	if rel == "" {
		return ""
	}
	if !strings.HasSuffix(rel, ".git") {
		rel += ".git"
	}
	return filepath.Join(BaseDir, rel)
}

// pathExists renvoie true si le chemin existe.
func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// initBare initialise un dépôt bare si nécessaire.
func initBare(repoPath string) error {
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		return err
	}
	// Crée un dépôt nu si pas déjà créé
	if pathExists(repoPath) {
		return nil
	}
	_, err := git.PlainInit(repoPath, true)
	return err
}

// ----- API -----

type createRepoReq struct {
	Name string `json:"name"`
}

// createRepoHandler crée un dépôt bare à la demande (pas d'auto-création ailleurs).
func createRepoHandler(c *fiber.Ctx) error {
	var req createRepoReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("JSON invalide")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("name requis")
	}
	if !strings.HasSuffix(name, ".git") {
		name += ".git"
	}
	repoPath := filepath.Join(BaseDir, filepath.Clean(strings.TrimPrefix(name, "/")))
	if pathExists(repoPath) {
		return c.Status(fiber.StatusConflict).SendString("dépôt existe déjà")
	}
	// Pour le moment, seule la création FS est supportée
	if StorageBackend != "fs" {
		return c.Status(fiber.StatusNotImplemented).SendString("création non supportée pour backend: " + StorageBackend)
	}
	if err := initBare(repoPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"name": name})
}

// writeServiceAdvertisement encode la première pkt-line spéciale
// "# service=git-<service>\n" suivie d'un flush-pkt, selon le protocole Smart HTTP.
func writeServiceAdvertisement(w io.Writer, service string) error {
	payload := []byte("# service=" + service + "\n")
	n := 4 + len(payload)
	// encodage longueur pkt-line sur 4 hex digits
	hdr := []byte{
		byte("0123456789abcdef"[(n>>12)&0xF]),
		byte("0123456789abcdef"[(n>>8)&0xF]),
		byte("0123456789abcdef"[(n>>4)&0xF]),
		byte("0123456789abcdef"[n&0xF]),
	}
	if _, err := w.Write(hdr); err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	// flush-pkt
	_, err := w.Write([]byte("0000"))
	return err
}

// ----- SSH Server -----

// startSSHServer démarre un serveur SSH qui gère git-upload-pack et git-receive-pack.
func startSSHServer(addr string) error {
	srv, err := NewGitSSHServer(addr)
	if err != nil {
		return err
	}
	return srv.ListenAndServe()
}

// NewGitSSHServer construit un serveur SSH Git prêt à être servi.
func NewGitSSHServer(addr string) (*gliderssh.Server, error) {
	hostKeyPath := filepath.Join(BaseDir, "ssh_host_key")
	signer, err := ensureHostKey(hostKeyPath)
	if err != nil {
		return nil, err
	}
	srv := &gliderssh.Server{
		Addr: addr,
		// Très permissif pour un exemple: accepte tout (password ou clé publique)
		PasswordHandler:  func(ctx gliderssh.Context, pass string) bool { return true },
		PublicKeyHandler: func(ctx gliderssh.Context, key gliderssh.PublicKey) bool { return true },
		Handler:          handleSSHSession,
	}
	srv.AddHostKey(signer)
	return srv, nil
}

// ensureHostKey charge/génère une clé hôte ed25519 et retourne un Signer utilisable par SSH.
func ensureHostKey(path string) (gossh.Signer, error) {
	if data, err := os.ReadFile(path); err == nil {
		if block, _ := pem.Decode(data); block != nil {
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				if sk, ok := key.(ed25519.PrivateKey); ok {
					return gossh.NewSignerFromKey(sk)
				}
			}
		}
	}
	// Génère une nouvelle clé et sauvegarde
	_, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	// Sauvegarde en PKCS#8 PEM
	b, err := x509.MarshalPKCS8PrivateKey(sk)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: b}); err != nil {
		return nil, err
	}
	return gossh.NewSignerFromKey(sk)
}

// handleSSHSession gère une session SSH entrante et exécute le service Git demandé.
func handleSSHSession(s gliderssh.Session) {
	cmd := s.RawCommand()
	service, repoArg := parseGitCommand(cmd)
	if service == "" || repoArg == "" {
		_, _ = io.WriteString(s.Stderr(), "commande git invalide\n")
		_ = s.Exit(1)
		return
	}

	repoPath := repoPathFromSSHArg(repoArg)
	if repoPath == "" {
		_, _ = io.WriteString(s.Stderr(), "dépôt invalide\n")
		_ = s.Exit(1)
		return
	}

	// Prépare le serveur go-git
	if service == "git-receive-pack" && !pathExists(repoPath) {
		if err := initBare(repoPath); err != nil {
			_, _ = io.WriteString(s.Stderr(), "échec init dépôt: "+err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
	}
	if !pathExists(repoPath) {
		_, _ = io.WriteString(s.Stderr(), "dépôt introuvable\n")
		_ = s.Exit(1)
		return
	}

	srv, ep, err := repoServer(repoPath)
	if err != nil {
		_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	switch service {
	case "git-upload-pack":
		up, err := srv.NewUploadPackSession(ep, nil)
		if err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		// En SSH, on envoie d'abord les refs annoncées
		adv, err := up.AdvertisedReferences()
		if err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		if err := adv.Encode(s); err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		// Puis on lit la requête du client et on répond
		req := packp.NewUploadPackRequest()
		if err := req.Decode(s); err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		resp, err := up.UploadPack(context.Background(), req)
		if err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		defer resp.Close()
		if err := resp.Encode(s); err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		_ = s.Exit(0)
	case "git-receive-pack":
		rp, err := srv.NewReceivePackSession(ep, nil)
		if err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		// Envoie des refs annoncées
		adv, err := rp.AdvertisedReferences()
		if err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		if err := adv.Encode(s); err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		// Lit la requête de mise à jour et répond avec le report-status
		req := packp.NewReferenceUpdateRequest()
		if err := req.Decode(s); err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		report, err := rp.ReceivePack(context.Background(), req)
		if err != nil {
			// Même en cas d'erreur, tente d'encoder le report si possible
			_ = report.Encode(s)
			_ = s.Exit(1)
			return
		}
		if err := report.Encode(s); err != nil {
			_, _ = io.WriteString(s.Stderr(), err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		_ = s.Exit(0)
	default:
		_, _ = io.WriteString(s.Stderr(), "service non supporté\n")
		_ = s.Exit(1)
	}
}

// parseGitCommand extrait le service et l'argument dépôt depuis une commande SSH.
// Ex: "git-upload-pack '/demo.git'" -> ("git-upload-pack", "/demo.git")
func parseGitCommand(cmd string) (service, repoArg string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "", ""
	}
	// Les commandes sont de la forme: git-upload-pack 'path' ou git-receive-pack 'path'
	if strings.HasPrefix(cmd, "git-upload-pack") {
		return "git-upload-pack", strings.TrimSpace(strings.TrimPrefix(cmd, "git-upload-pack"))
	}
	if strings.HasPrefix(cmd, "git-receive-pack") {
		return "git-receive-pack", strings.TrimSpace(strings.TrimPrefix(cmd, "git-receive-pack"))
	}
	return "", ""
}

// repoPathFromSSHArg nettoie l'argument dépôt SSH ('/demo.git') et retourne le chemin local.
func repoPathFromSSHArg(arg string) string {
	arg = strings.TrimSpace(arg)
	// Supprime guillemets simples/doubles et éventuel leading ':'
	arg = strings.Trim(arg, "'\"")
	arg = strings.TrimPrefix(arg, ":")
	// Supprime éventuel host: et leading slash
	if i := strings.Index(arg, ":"); i >= 0 {
		arg = arg[i+1:]
	}
	arg = strings.TrimPrefix(arg, "/")
	if arg == "" {
		return ""
	}
	if !strings.HasSuffix(arg, ".git") {
		arg += ".git"
	}
	return filepath.Join(BaseDir, arg)
}
