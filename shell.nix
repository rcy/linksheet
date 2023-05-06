with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    go
    sqlite
    flyctl
    google-cloud-sdk
  ];
}
