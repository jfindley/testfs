# TestFS

TestFS implements a full POSIX filesystem in memory in Go, similar to a basix Linux/UNIX filesystem.

It's designed primarily for running detailed tests, and as such it also provides OSFS which wraps the normal core "os" commands.

All syntax and functionality should be identical to the core "os" package.

To use this in your projects, create a Filesystem variable, and use either NewOSFS or NewTestFS to use either the normal on-disk filesystem
or the in-memory TestFS filesystem.
NewTestFs takes two parameters, the root UID and GID.  To use the current user, there's a helper function: NewLocalTestFS.

# Stability

Current status of the code is first alpha, at best.  There may well be bugs.  However, the interface is set in stone, as it is designed to exactly match the core "os" package.  Any behaviour that doesn't match "os" is a bug, and will be fixed.

# Performance

It's entirely in RAM, so general IO performance is excellent.  However, making this behave like a real POSIX filesystem introduces a bunch of overheads in areas like directory creation and traversal, for example.  A well implemented key/value store will be considerably faster, albeit with far fewer features.

# Can it do XYZ?

If XYZ is a common feature of POSIX filesystems, yes.  This supports xattrs (with no space limit), so testing selinux contexts etc should all be fine.

You cannot, however, use this to run external applications in memory without modifying the application to link against TestFS.
