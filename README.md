# SOPS pre-commit

This tool is designed to make git pre-commit hooks for validating sops encryption in a respository easier. It accepts
a list of modified files or a pipe from stdin.  It will source a .sops.yaml config file from the current working directory in the same fashion as sops.
If a sops config is found, it will determine which files in the change set match creation_rules from config and attempt to decrypt them into memory to validate
encryption. If there is no sops configuration, it is assumed that the list of files is prefiltered to only include sops encrypted files and all files are validated via
decryption.

Decryption is the most effective validation tool for an encrypted file and the author would already need to have access to the encryption keys to create the change set.



## Demo

1. Build the tool and import the test key.

```
make build
cd example
gpg --import ./infra_test_key.asc
cat .sops.yaml
```

2. Create a new file in the secrets directory.

```
sops secrets/new_file.yaml
```

3. Add the new encrypted file to the cached changes

```
git add secrets/new_file.yaml
```

4. Pass the change set to the tool.

```
cd $(git rev-parse --show-toplevel)
git diff --name-only --cached | ./sopsprecommit

cd -

```
