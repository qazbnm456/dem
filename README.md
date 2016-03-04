# Docker Environment Manager

Docker Environment Manager enables the ability to change environment smoothly for your Docker server. Project inspired by [rvm](https://rvm.io/) and [dvm](https://github.com/getcarina/dvm).
This tool modifies your docker default configuration and restart your docker service to switch between different Docker's storage base, see [here](https://forums.docker.com/t/how-do-i-change-the-docker-image-installation-directory/1169) for details.

## Notice
This tool is still under heavy development.

## Installation
1. Clone this repo.
2. go install.

**Currently only support Ubuntu/Debian**

## Usage

To create the imgset, do this:

    dem create [<imgset>]

Now in any new shell use the installed imgset:

    dem use [<imgset>]

If you want to see what imgsets are installed:

    dem list

