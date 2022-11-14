# Testing

Tests on Windows depend on a connected virtual COM port pair being available at COM 5 and 6. The author uses 
com0com [https://com0com.sourceforge.net/] to create this port pair. Accordingly, tests on Windows have to
be run with the `-p 1` flag.

Tests on Linux depend on `socat`. Ensure that the binary is available in your path.