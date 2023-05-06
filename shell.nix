with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    go
    gopls
    sqlite
    flyctl
    google-cloud-sdk
  ];
}
