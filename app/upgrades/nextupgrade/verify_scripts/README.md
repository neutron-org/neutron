# Upgrade verification scripts

1. run fetch_pre_mig_state.sh specifying a remote node if needed. It will create a `pre_mig_state.json` file in the directory of the script and populate it with upgrade sensitive network state values.

```sh
NODE=http://37.27.66.241:26657 ./fetch_pre_mig_state.sh
```

2. perform the upgrade on the node

3. run fetch_post_mig_state.sh specifying a remote node if needed. It will create a `post_mig_state.json` file in the directory of the script and populate it with upgrade sensitive network state values.

```sh
NODE=http://37.27.66.241:26657 ./fetch_post_mig_state.sh
```

4. run verigy_mig.sh. It will use the two .json files and perform verification checks.

```sh
./verify_mig.sh
```
