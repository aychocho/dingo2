package dingo

import (
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// remoteFileSystem implements the FileSystem interface
type remoteFileSystem struct {
	client *ssh.Client
	sftp   *sftp.Client
	config *SftpConfig
	err    error
}

/*
* Reads the entire contents of a remote file and returns it as bytes
* Inputs: name (string) - path to the remote file to read
* Outputs: []byte containing file contents, error if file cannot be read or SFTP error exists
 */
func (rfs *remoteFileSystem) ReadFile(name string) ([]byte, error) {
	if rfs.err != nil {
		return nil, rfs.err
	}

	f, err := rfs.sftp.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

/*
* Writes byte data to a remote file with specified permissions, creating or truncating as needed
* Inputs: name (string) - remote file path, data ([]byte) - content to write, perm (os.FileMode) - file permissions
* Outputs: error if file cannot be written or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	if rfs.err != nil {
		return rfs.err
	}

	f, err := rfs.sftp.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

/*
* Transfers a file from local filesystem to remote server via SFTP
* Inputs: localPath (string) - path to local source file, remotePath (string) - destination path on remote server
* Outputs: error if transfer fails due to file access or network issues, nil on successful transfer
 */
func (rfs *remoteFileSystem) Upload(localPath, remotePath string) error {
	if rfs.err != nil {
		return rfs.err
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFile, err := rfs.sftp.OpenFile(remotePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	return err
}

/*
* Transfers a file from remote server to local filesystem via SFTP
* Inputs: remotePath (string) - path to remote source file, localPath (string) - destination path on local filesystem
* Outputs: error if transfer fails due to file access or network issues, nil on successful transfer
 */
func (rfs *remoteFileSystem) Download(remotePath, localPath string) error {
	if rfs.err != nil {
		return rfs.err
	}

	remoteFile, err := rfs.sftp.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	return err
}

/*
* Creates a single directory on the remote server (parent directories must exist)
* Inputs: path (string) - remote directory path to create
* Outputs: error if directory creation fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) Mkdir(path string) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.Mkdir(path)
}

/*
* Creates a directory and all necessary parent directories on the remote server
* Inputs: path (string) - remote directory path to create (including parents)
* Outputs: error if directory creation fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) MkdirAll(path string) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.MkdirAll(path)
}

/*
* Removes a file or empty directory from the remote server
* Inputs: path (string) - remote path to file or empty directory to remove
* Outputs: error if removal fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) Remove(path string) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.Remove(path)
}

/*
* Removes a directory from the remote server (directory must be empty)
* Inputs: path (string) - remote directory path to remove
* Outputs: error if directory removal fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) RemoveDirectory(path string) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.RemoveDirectory(path)
}

/*
* Returns file information for a remote path, following symbolic links
* Inputs: path (string) - remote file or directory path to examine
* Outputs: os.FileInfo containing file metadata, error if stat fails or SFTP error exists
 */
func (rfs *remoteFileSystem) Stat(path string) (os.FileInfo, error) {
	if rfs.err != nil {
		return nil, rfs.err
	}
	return rfs.sftp.Stat(path)
}

/*
* Returns file information for a remote path without following symbolic links
* Inputs: path (string) - remote file or directory path to examine
* Outputs: os.FileInfo containing file metadata, error if lstat fails or SFTP error exists
 */
func (rfs *remoteFileSystem) Lstat(path string) (os.FileInfo, error) {
	if rfs.err != nil {
		return nil, rfs.err
	}
	return rfs.sftp.Lstat(path)
}

/*
* Reads a remote directory and returns information about all contained files and subdirectories
* Inputs: path (string) - remote directory path to read
* Outputs: []os.FileInfo slice containing metadata for directory contents, error if read fails or SFTP error exists
 */
func (rfs *remoteFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	if rfs.err != nil {
		return nil, rfs.err
	}
	return rfs.sftp.ReadDir(path)
}

/*
* Changes the permissions of a remote file or directory
* Inputs: path (string) - remote file or directory path, mode (os.FileMode) - new permissions to set
* Outputs: error if permission change fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) Chmod(path string, mode os.FileMode) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.Chmod(path, mode)
}

/*
* Changes the ownership of a remote file or directory
* Inputs: path (string) - remote file or directory path, uid (int) - new user ID, gid (int) - new group ID
* Outputs: error if ownership change fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) Chown(path string, uid, gid int) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.Chown(path, uid, gid)
}

/*
* Renames or moves a remote file or directory from old path to new path
* Inputs: oldname (string) - current remote path, newname (string) - new remote path
* Outputs: error if rename/move fails or SFTP error exists, nil on success
 */
func (rfs *remoteFileSystem) Rename(oldname, newname string) error {
	if rfs.err != nil {
		return rfs.err
	}
	return rfs.sftp.Rename(oldname, newname)
}

/*
* Closes the SFTP session and releases associated resources
* Inputs: none
* Outputs: error if SFTP session close fails or SFTP error exists, nil on successful close
 */
func (rfs *remoteFileSystem) Close() error {
	if rfs.err != nil {
		return rfs.err
	}
	if rfs.sftp != nil {
		return rfs.sftp.Close()
	}
	return nil
}
