# Linux

Pathvector can be installed on most Linux (and other UNIX-based) systems.

* Install pathvector from a [release](https://github.com/natesales/pathvector/releases) or add the [repo](https://github.com/natesales/repo) and install the `pathvector` deb/rpm package.
* Create the runtime directory `sudo mkdir -p /var/run/pathvector/cache/`
* Write your configuration to `/etc/pathvector.yml`
* Add pathvector to your crontab or other scheduler `0 */12 * * * pathvector`
* Run `pathvector` to validate your config
