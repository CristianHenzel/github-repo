[![GoDoc](https://godoc.org/github.com/CristianHenzel/github-repo?status.svg)](https://godoc.org/github.com/CristianHenzel/github-repo)
[![Build Status](https://travis-ci.org/CristianHenzel/github-repo.svg?branch=master)](https://travis-ci.org/CristianHenzel/github-repo)
[![Go Report Card](https://goreportcard.com/badge/github.com/CristianHenzel/github-repo)](https://goreportcard.com/report/github.com/CristianHenzel/github-repo)
[![License: LGPL v2.1](https://img.shields.io/badge/License-LGPL%20v2.1-blue.svg)](https://www.gnu.org/licenses/lgpl-2.1)

----

gr is a GitHub repository management tool

**WARNING!** gr stores all data (including your token) in a plaintext file. If you consider this a security issue or if you are sharing your machine with other people, do not use this tool!

# Installation
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
gr init -u USERNAME -t TOKEN
```
or, if you are using GitHub Enterprise:
```
gr init -u USERNAME -t TOKEN -r https://example.com/api/v3/
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
