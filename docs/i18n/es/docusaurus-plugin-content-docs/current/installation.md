---
sidebar_position: 2
---

# Instalación

Todas las versiones de Pathvector para todas las plataformas soportadas están disponibles para su descarga en la
última [página de versiones](https://github.com/natesales/pathvector/releases).

Se recomienda ejecutar Pathvector cada 12 horas para actualizar las listas de prefijos de IRR y los límites de prefijos desde PeeringDB.
Añadiendo `0 */12 * * * pathvector generate` a su crontab que actualizará los filtros a las 12 AM y PM de cada día.

La única dependencia que necesaria es `bird >= 2.0.7`, pero algunas funciones se requiere dependencias adicionales:

- Filtrado de RPKI: Servidor de RTR como [gortr](https://github.com/cloudflare/gortr) o el servidor RTR público de Cloudflare a `rtr.rpki.cloudflare.com:8282`
- Generación de la lista de prefijos IRR: [bgpq4](https://github.com/bgp/bgpq4)
- Daemon de VRRP: [keepalived](https://github.com/acassen/keepalived)

## Repositorio de Paquetes

Las versiones del Pathvector >= 5.1.2 están disponibles en el repositorio de paquetes https://repo.pathvector.io. Los paquetes también están en
[repo de natesales](https://github.com/natesales/repo) para que sean compatibles con las instalaciones existentes, pero por seguridad se recomienda utilizar el
repo.pathvector.io para todas las nuevas instalaciones debido a la mayor seguridad por las firmas GPG. Los paquetes en repo.pathvector.io están firmados
con [`0983 AC66 7B4F 0B54 F69D`](https://repo.pathvector.io/pgp.asc). Tenga en cuenta que los paquetes descargados de GitHub no están firmados.

Pathvector en Linux está disponible para amd64, aarch64 y mips64 como binarios y paquetes deb y rpm desde [releases](https://github.com/natesales/pathvector/releases).

## Debian

```shell
curl https://repo.pathvector.io/pgp.asc > /usr/share/keyrings/pathvector.asc
echo "deb [signed-by=/usr/share/keyrings/pathvector.asc] https://repo.pathvector.io/apt/ stable main" > /etc/apt/sources.list.d/pathvector.list
apt update && apt install -y pathvector
```

## CentOS

```shell
yum install -y yum-utils
yum-config-manager --add-repo https://repo.pathvector.io/yum/pathvector.repo
yum install pathvector
```

## VyOS

VyOS está basado en Debian, see [Debian](#debian) para más información.

## TNSR

TNSR está basado en CentOS, see [CentOS](#centos) para más información.

## FreeBSD

Pathvector is available as an amd64/aarch64 binary for FreeBSD from
the [releases](https://github.com/natesales/pathvector/releases) page.

## Ubiquiti EdgeOS

Los EdgeRouters de Ubiquiti están basados en Debian. Pathvector soporta los routers ER-8-XG, ER-12P, ER-12, ERPro-8, ER-8, ER-6P, ERPoe-5,
ER-4 y ERLite-3 basados en la arquitectura MIPS64. Mira [Debian](#debian) para más información.

No se admite (MIPS1004Kc): ER-X, ER-X-SFP, ER-10X

## Arista EOS

Pathvector puede funcionar en los conmutadores Arista instalando un archivo SWIX (Switch Extension)
de [lanzamientos](https://github.com/natesales/pathvector/releases). Primero, copie el archivo do `pathvector-arista.swix` a `/mnt/flash/` en tu conmutador Arista.

Desde el EOS CLI:

1. Copiar la extensión en extensiones: `copy flash:pathvector-arista.swix extensions:`
2. Instalar la extensión: `extension pathvector-arista.swix`
3. Marque la extensión a instalar en el arranque: `copy installed-extensions boot-extensions`
4. Añade el alias pathvector: `alias pathvector "bash pathvector"`
5. Añadir Pathvector al planificador (opcional): `schedule pathvector interval 720 max-log-files 1 command pathvector -c /mnt/flash/pathvector.yml -o /mnt/flash/bird/ -s /run/bird.ctl`
6. Guarda los cambios: `wr mem`

## Juniper JunOS

Pathvector puede ejecutarse en dispositivos Juniper mediante la instalación de un paquete de extensión firmado de JunOS. En primer lugar, descargue la versión de Pathvector
Juniper, ya sea en su máquina local y SCP sobre el router / conmutador de Juniper, o usa de él directamente en
el comando `request system software` sustituyendo el nombre del archivo por la URL del mismo.

```shell
junos# set system extensions providers pathvector license-type customer deployment-scope commercial
junos> request system software add pathvector-juniper.tgz
```

## Cisco IOx

Pathvector puede funcionar en dispositivos Cisco compatibles con IOx mediante la instalación de una versión del paquete IOx, o directamente en el dispositivo con
ioxclient. See https://developer.cisco.com/docs/iox/#!app-management/application-management para más información.

## Nokia Service Router (SR) Linux

Nokia SR Linux está basado en CentOS, see [CentOS](#centos) para más información.

## Mikrotik RouterOS

Pathvector puede ser instalado en >= RouterOS v7.1rc3. Compruebe
el [Mikrotik Product Matrix](https://mikrotik.com/products/matrix) para el listado de hardware más reciente y
el [RouterOS Container](https://help.mikrotik.com/docs/display/ROS/Container) referencia para las instrucciones de instalación
para el paquete de contenedores.

To build a Docker image for an alternate architecture:

```bash
docker build --output type=tar,dest=pathvector-mikrotik-arm64v8.tar -t pathvector-cron:arm64v8 --build-arg ARCH=arm64v8 -f ../vendorbuild/mikrotik/Dockerfile ..
```

## Construir desde el origen

Pathvector puede construirse fácilmente a partir del código fuente para algunos, pero no todos, los
el [muchas plataformas compatibles con Go](https://github.com/golang/go/blob/master/src/go/build/syslist.go).

Por ejemplo, para construir Pathvector para Macs M1:

```bash
git clone https://github.com/natesales/pathvector && cd pathvector
GOOS=darwin GOARCH=arm64 go build
```
