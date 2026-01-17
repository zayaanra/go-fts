package peer

import (
	"encoding/binary"
	"io"
	"net"
	"os"

	"github.com/zayaanra/go-fts/api"
	"github.com/zayaanra/go-fts/crypt"
)

// TODO: How can we make handle.go cleaner and more idiomatic?

func handleAck(p *Peer) error {
	if p.Role == PAKE_INITIATOR {
		err := p.Conn.WriteJSON(api.Message{
			Protocol:  api.SHARE_PUBLIC_KEY,
			SessionID: p.Session.ID,
			Data: p.Curve.Bytes(),
		})
		return err
	}
	return nil
}

func handleSharePublicKey(p *Peer, publicKey []byte) error {
	err := p.Curve.Update(publicKey)
	if err != nil {
		return err
	}

	if p.Role == PAKE_RESPONDER {
		err = p.Conn.WriteJSON(api.Message{
			Protocol:  api.SHARE_PUBLIC_KEY,
			SessionID: p.Session.ID,
			Data: p.Curve.Bytes(),
		})
		if err != nil {
			return err
		}
	}

	sessionKey, _ := p.Curve.SessionKey()
	p.Session.Key = crypt.HKDF(sessionKey)

	if p.Role == PAKE_RESPONDER {
		ln, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return err
		}

		addr := ln.Addr().(*net.TCPAddr).String()
		encrypted, err := crypt.EncryptAES([]byte(addr), p.Session.Key)
		if err != nil {
			return err
		}

		err = p.Conn.WriteJSON(
			api.Message{
				Protocol:  api.SHARE_IP,
				SessionID: p.Session.ID,
				Data:      encrypted,
			},
		)
		if err != nil {
			return err
		}

		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()

		p.SenderIP = conn.RemoteAddr().String()

		file, err := os.Create("result.txt")
		if err != nil {
			return err
		}
		defer file.Close()

		for {
			var lenBuf [8]byte
			_, err := io.ReadFull(conn, lenBuf[:])
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			chunkLen := binary.BigEndian.Uint64(lenBuf[:])

			encrypted := make([]byte, chunkLen)
			_, err = io.ReadFull(conn, encrypted)
			if err != nil {
				return err
			}

			// TODO: Add/update progress bar here

			decrypted, err := crypt.DecryptAES(encrypted, p.Session.Key)
			if err != nil {
				return err
			}
			
			_, err = file.Write(decrypted)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func handleShareIP(p *Peer, data []byte) error {
	decrypted, err := crypt.DecryptAES(data, p.Session.Key)
	if err != nil {
		return err
	}

	addr := string(decrypted)
	p.ReceiverIP = addr
	return nil
}