package dingo

import (
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
)

/*
* Establishes an SSH connection using username and password authentication
* Inputs: addr (string) - SSH server address with port, user (string) - username, password (string) - user password
* Outputs: SSHClient interface implementation, error if connection fails
 */
func ConnectWithPassword(addr, user, password string) (SSHClient, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return connectWithConfig("tcp", addr, config)
}

/*
* Establishes an SSH connection using SSH private key authentication
* Inputs: addr (string) - SSH server address with port, user (string) - username, keyPath (string) - path to private key file
* Outputs: SSHClient interface implementation, error if connection or key parsing fails
 */
func ConnectWithKey(addr, user, keyPath string) (SSHClient, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return connectWithConfig("tcp", addr, config)
}

/*
* Establishes an SSH connection using SSH private key with passphrase authentication
* Inputs: addr (string) - SSH server address with port, user (string) - username, keyPath (string) - path to private key file, passphrase (string) - key passphrase
* Outputs: SSHClient interface implementation, error if connection or key parsing fails
 */
func ConnectWithKeyAndPassphrase(addr, user, keyPath, passphrase string) (SSHClient, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return connectWithConfig("tcp", addr, config)
}

/*
* Establishes an SSH connection using a custom SSH client configuration
* Inputs: addr (string) - SSH server address with port, config (*ssh.ClientConfig) - custom SSH configuration
* Outputs: SSHClient interface implementation, error if connection fails
 */
func ConnectWithConfig(addr string, config *ssh.ClientConfig) (SSHClient, error) {
	return connectWithConfig("tcp", addr, config)
}

/*
* Establishes an SSH connection using an existing network connection
* Inputs: conn (net.Conn) - existing network connection, addr (string) - server address, config (*ssh.ClientConfig) - SSH configuration
* Outputs: SSHClient interface implementation, error if SSH handshake fails
 */
func ConnectWithConnection(conn net.Conn, addr string, config *ssh.ClientConfig) (SSHClient, error) {
	ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}

	sshClient := ssh.NewClient(ncc, chans, reqs)
	return newClient(sshClient, DefaultClientConfig), nil
}

/*
* Helper function that establishes connections using the specified network type and SSH configuration
* Inputs: network (string) - network type (usually "tcp"), addr (string) - server address, config (*ssh.ClientConfig) - SSH configuration
* Outputs: SSHClient interface implementation, error if connection fails
 */
func connectWithConfig(network, addr string, config *ssh.ClientConfig) (SSHClient, error) {
	sshClient, err := ssh.Dial(network, addr, config)
	if err != nil {
		return nil, err
	}

	return newClient(sshClient, DefaultClientConfig), nil
}

/*
* Returns a host key callback for secure host key verification (currently returns insecure callback as placeholder)
* Inputs: knownHostsFile (string) - path to known_hosts file for verification
* Outputs: ssh.HostKeyCallback function, error if unable to create callback
 */
func SecureHostKeyCallback(knownHostsFile string) (ssh.HostKeyCallback, error) {
	// In a real implementation, this would verify against known_hosts
	// For now, return the insecure version with a TODO
	// TODO: Implement proper host key verification
	return ssh.InsecureIgnoreHostKey(), nil
}
