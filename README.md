# SafeBox

A Linux-only utility to mount Truecrypt/Veracrypt containers along with special folders like `.ssh` and `.gnupg`.

## Prerequisites

1. Make sure that all extFAT drivers are installed.<br>
    On Ubuntu it means you may need to install the following packages:

    ```
    $ sudo apt install exfat-fuse exfat-utils
    ```

    On Fedora 30+ it can be done by downloading a file [here](https://www.rpmfind.net/linux/rpm2html/search.php?query=fuse-exfat) and applying the command:

    ```
    $ sudo dnf install fuse-exfat-<the-latest-version>.x86_64.rpm
    ```

2. Install the console version of [Veracrypt](https://veracrypt.fr/en/Downloads.html).

## Installation

```
GO111MODULE=on go install github.com/pkorotkov/safebox/cmd/safebox@latest
```
