# Go-FTS (File Transfer Service)
**Go-FTS** is a command line utility that acts a file transfer service written in Go. The CLI tool is easy and simple to use. It simply requires that two machines have the binary installed ready to use.

## Features
- Send a file to any computer
- Receive a file from the sending computer

## Tech Stack
- Go

## Implementation
The project works by allowing two peers (sender and receiver) to communicate with each other. A sender and receiver initially know nothing about each other and in order for them to find each other, a process known as **password authenticated key exchange (PAKE)** is used. This process involves generating a short passphrase and allowing both sides to generate a private key using it. Before the key is generated, they communicate via a mailbox server (the mailbox server simply pairs senders and receivers with each other). After both peers are paired, they begin generating their public keys. With the help of the passphrase and their public keys, both peers generate a private key that is used after running through a key derivation function (**HKDF**) for encryption/decryption.

After **PAKE** is completed, both peers exchange connection information (their IP addresses) so that the sender can directly open a TCP connection to the receiver. The file data is then streamed to the receiver.

## Installation
### Option 1:
1. Ensure that the latest version of `go` is installed.
2. Clone the repository to somewhere on your machine: `git clone git@github.com:zayaanra/go-fts.git`
3. You can run `go run main.go <args>` where `<args>` are the arguments to the program.
### Option 2:
1. Download the binary and run `./fts`.

## Usage
- The *send* command takes in two arguments: the IP address of the receiving machine and the path to the file you want to send.
	- `./fts send /path/to/input/`
- The *receive* command takes in one argument: the path to where the received file should be written.
	- `./fts receive /path/to/output/`
