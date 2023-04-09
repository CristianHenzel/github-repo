[![CI](https://github.com/CristianHenzel/github-repo/actions/workflows/ci.yml/badge.svg)](https://github.com/CristianHenzel/github-repo/actions/)
[![Go Report](https://goreportcard.com/badge/github.com/CristianHenzel/github-repo)](https://goreportcard.com/report/github.com/CristianHenzel/github-repo)
[![Maintainability](https://img.shields.io/codeclimate/maintainability-percentage/CristianHenzel/github-repo.svg)](https://codeclimate.com/github/CristianHenzel/github-repo/maintainability)
[![License: LGPL v2.1](https://img.shields.io/github/license/CristianHenzel/github-repo.svg?color=blue)](https://www.gnu.org/licenses/lgpl-2.1)

----

gr is a GitHub repository management tool

**WARNING!** gr stores all data (including your token) in a plaintext file. If you consider this a security issue or if you are sharing your machine with other people, do not use this tool!

# Installation
### Using prebuilt binaries:
```
curl -L https://github.com/CristianHenzel/github-repo/releases/latest/download/gr_linux_amd64 -o /usr/local/bin/gr
chmod 0755 /usr/local/bin/gr
```
### Building from source:
Prerequisites: git, golang, make, upx

To install prerequisites on debian/ubuntu, run:
```
sudo apt-get install git golang make upx-ucl
```

To install gr:
```
git clone https://github.com/CristianHenzel/github-repo
cd github-repo
make install
```

# Usage
First, create the configuration:
```
gr init -u USERNAME -t TOKEN -d SOMEDIR -c 10
```
or, if you are using GitHub Enterprise:
```
gr init -c 10 -u USERNAME -t TOKEN -r https://example.com/api/v3/ -d SOMEDIR -e "repo1|SOMEORG/repo-.*" -s
```

After the configuration is created, you can pull all repositories using:
```
gr pull
```

you can view the status of the repositories using:
```
gr status
```

and you can push all repositories using:
```
gr push
```

After creating new repositories on the server or after user data changes, you can update the local configuration using:
```
gr update
```
