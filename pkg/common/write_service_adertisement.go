package common

import "io"

func WriteServiceAdvertisement(w io.Writer, service string) error {
	payload := []byte("# service=" + service + "\n")
	n := 4 + len(payload)
	// encode pkt-line length as 4 hex digits
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
