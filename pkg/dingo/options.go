package dingo

// SFTP Option functions - simplified for single-threaded operation

/*
* Creates an SFTP option that sets the maximum packet size for file transfer operations
* Inputs: size (int) - maximum packet size in bytes for SFTP operations
* Outputs: SftpOption function that applies the packet size configuration
 */
func WithMaxPacket(size int) SftpOption {
	return func(config *SftpConfig) {
		config.MaxPacket = size
	}
}

/*
* Creates an SFTP option that enables or disables the use of fstat for file operations
* Inputs: enabled (bool) - whether to use fstat for file operations
* Outputs: SftpOption function that applies the fstat configuration
 */
func WithFstat(enabled bool) SftpOption {
	return func(config *SftpConfig) {
		config.UseFstat = enabled
	}
}

// Common SFTP option presets for single-threaded operation

/*
* Returns a set of SFTP options optimized for maximum transfer speed in single-threaded mode
* Inputs: none
* Outputs: []SftpOption slice containing speed-optimized configuration options
 */
func FastSftpOptions() []SftpOption {
	return []SftpOption{
		WithMaxPacket(65536), // Larger packets for speed
		WithFstat(false),     // Skip fstat for speed
	}
}

/*
* Returns a set of SFTP options optimized for safe and reliable file operations
* Inputs: none
* Outputs: []SftpOption slice containing safety-optimized configuration options
 */
func SafeSftpOptions() []SftpOption {
	return []SftpOption{
		WithMaxPacket(32768), // Smaller packets for reliability
		WithFstat(true),      // Use fstat for safety
	}
}

/*
* Returns a set of balanced SFTP options suitable for most use cases
* Inputs: none
* Outputs: []SftpOption slice containing balanced configuration options
 */
func DefaultSftpOptions() []SftpOption {
	return []SftpOption{
		WithMaxPacket(32768), // Balanced packet size
		WithFstat(false),     // Skip fstat for moderate speed
	}
}
