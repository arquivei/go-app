# uConfig

This is a slightly altered version of the `https://github.com/omeid/uconfig/`.

All credits for this module should go to the original creator.

Changes made in this package:

- Removed go.mod and go.sum files;
- Renamed packages;
- Removed examples directory;
- Changed `github.com/google/go-cmp/cmp` to `github.com/stretchr/testify` to reduce dependencies;
- Added a formatted usage output to print the configuration in more useful formats for easy copy and pasting;
