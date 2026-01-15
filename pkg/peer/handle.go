package peer

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"

	"github.com/zayaanra/go-fts/pkg/api"
	"github.com/zayaanra/go-fts/pkg/crypt"
)

func handleAck(p *Peer) error {
	if p.Role == PAKE_INITIATOR {
		err := p.Conn.WriteJSON(api.Message{
			Protocol: api.SHARE_PUBLIC_KEY,
			SessionID: p.Session.ID,
			PublicKey: p.Curve.Bytes(),
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
			Protocol: api.SHARE_PUBLIC_KEY,
			SessionID: p.Session.ID,
			PublicKey: p.Curve.Bytes(),
		})
		if err != nil {
			return err
		}
	}
	
	sessionKey, _ := p.Curve.SessionKey()
	p.Session.Key = crypt.HKDF(sessionKey)

	if p.Role == PAKE_RESPONDER {
		// TODO: Should probably let OS decide on a random port
		ln, err := net.Listen("tcp", "localhost:8091")
		if err != nil {
			return err
		}

		addr := ln.Addr().(*net.TCPAddr).String()
		encrypted, err := crypt.EncryptAES([]byte(addr), p.Session.Key)
		if err != nil {
			return err
		}
		
		err = p.Conn.WriteJSON(api.Message{
			Protocol: api.SHARE_IP,
			SessionID: p.Session.ID,
			Data: encrypted,
		})
		if err != nil {
			return err
		}
		
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()

		p.SenderIP = conn.RemoteAddr().String()
		
		var size uint64
		binary.Read(conn, binary.BigEndian, &size)
		buf := make([]byte, size) // TODO: makeslice: len out of range
		io.ReadFull(conn, buf)
		
		// TODO: Write file back in receive.go, handle file naming
		decrypted, err := crypt.DecryptAES(buf, p.Session.Key)
		file := &api.File{}
		json.Unmarshal(decrypted, file)
		os.WriteFile("result.txt", file.Data, 0777)
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
	if p.Role == PAKE_INITIATOR {
		// encrypted, err := crypt.EncryptAES([]byte(p.FileData), p.Session.Key)
		// if err != nil {
		// 	return err
		// }

		// binary.Write(conn, binary.BigEndian, uint64(len(encrypted)))
		// _, err = conn.Write(encrypted)
		// if err != nil {
		// 	return err
		// }
		// conn.Close()
	}
	return nil
}