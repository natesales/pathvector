# Arista

Wireframe can be run on Arista switches by installing a SWIX (Switch Extension) file. The [arista](https://github.com/natesales/wireframe/tree/main/arista) directory contains packaging configuration to build a SWIX bundle extension containing Wireframe, BIRD2, GoRTR, and bgpq4. 

## Installation

To build the SWIX on your machine:

1. Download the latest .swix file from [releases](https://github.com/natesales/wireframe/releases), or build it manually with `cd arista && make`
2. Copy the resulting SWIX bundle extension to /mnt/flash/ on the switch

On the switch from the EOS CLI:

1. Copy the extension to extensions: `copy flash:wireframe-bundle.swix extensions:`
2. Install the extension: `extension wireframe-bundle.swix`
3. Mark the extension to be installed on boot: `copy installed-extensions boot-extensions`
4. Restart the EOS CLI: `bash sudo pkill Cli` and log back into the switch
5. Create the BIRD directory: `bash sudo mkdir /mnt/flash/bird/`
6. Write your wireframe config: `bash sudo nano /mnt/flash/wireframe.yml`
7. Run wireframe: `wireframe -c /mnt/flash/wireframe.yml -o /mnt/flash/bird/ --no-configure`
8. Restart bird: `bash sudo systemctl restart bird`
9. Add Wireframe to scheduler: `schedule test interval 720 max-log-files 1 command wireframe -c /mnt/flash/wireframe.yml -o /mnt/flash/bird/ -s /run/bird.ctl`
10. Save changes: `wr mem`

## Usage

After installing the bundle extension, your switch will have a few new EOS CLI commands: `birdc`, `bgpq4`, and `wireframe`. These are just wrappers for the binaries installed on the underlying Linux system. 
