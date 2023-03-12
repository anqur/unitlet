# unitlet

A [Virtual Kubelet](https://github.com/virtual-kubelet/virtual-kubelet) provider for running systemd units that just
works. This project is heavily inspired by [systemk](https://github.com/virtual-kubelet/systemk).

There is just a limited set of supported features, like just wrapping basic functionalities of `dbus` and `journal`,
exposing them via K8s API. That's because:

* I have a farm of no greater than 50 machines to manage, per cluster
* I have the full control of these machines, with 250+ GiB memory and 50+ TiB disks, for distributed storage
* I don't think Ansible-based automation is decently enough for administering
* I want to try some fancy yet battle-tested tooling

Todos:

* [ ] Fetching container logs
* [ ] Exposing metrics
* [ ] Ah yes, haven't tested all of these yet xD

## License

MIT
