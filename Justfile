build:
    go build
    sed -i "s/version:.*/version: 0.0.1-`git rev-parse --short HEAD`/g" nfpm.yaml
    nfpm package -p deb
