package utils

// Store the build verion information; set by linker flag.
//
// SleuthX fork: the release build (build-androidqf.yml) passed no -X, so -version printed an empty
// string and could not identify a build. The default now names this fork tag; a linker flag still
// overrides it.
var Version = "1.8.3-sleuthx3"
