# terraform-plan-summary

A CLI tool that **summarizes Terraform JSON plan files** in a human-readable, colored table format.  
Supports both **detailed** per-project views and **summary overviews** across multiple plan files.

---

## Features

- 🔍 Summarizes resource actions (`create`, `update`, `delete`, etc.)
- 🎨 Colorized table output.
- 📁 Handles multiple Terraform plan files across nested directories
- 🔎 Extracts environment/project using regex

---

## Installation


tf-plan-summary is a Go binary available for Linux, MacOS and Windows from the link:https://github.com/tf-plan-summary/tf-plan-summary/releases[release page].

**Verify File Checksum Signature**

Instead of signing all release assets, tf-plan-summary signs the checksums file containing the different release assets checksum.
You can download/copy the three files 'checksums.txt.pem', 'checksums.txt.sig', 'checksums.txt' from the latest https://github.com/tf-plan-summary/tf-plan-summary/releases/latest[release].
Once you have the three files locally, you can execute the following command

```
cosign verify-blob --certificate-identity-regexp "https://github.com/tf-plan-summary/tf-plan-summary" --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' --cert https://github.com/tf-plan-summary/tf-plan-summary/releases/download/v0.1.0/checksums.txt.pem --signature https://github.com/tf-plan-summary/tf-plan-summary/releases/download/v0.1.0/checksums.txt.sig checksums.txt
```

A successful output looks like

```
Verified OK
```


Now you can verify the assets checksum integrity.

**Verify File Checksum Integrity**

Before verifying the file integrity, you should first verify the checksum file signature.
Once you've download both the checksums.txt and your binary, you can verify the integrity of your file by running:

```
sha256sum --ignore-missing -c checksums.txt
```

