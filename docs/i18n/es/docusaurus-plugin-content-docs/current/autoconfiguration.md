---
title: Autoconfiguración
sidebar_position: 5
---

Pathvector puede construir ciertas configuraciones de pares automáticamente. El comando `match` encuentra los IXP comunes para un ASN y da como resultado un cuerpo de correo electrónico en texto plano o un fragmento YAML de Pathvector.

## Configuración de YAML

La configuración YAML generada tendrá una sangría de 2 espacios por defecto para ajustarse a la clave de los pares de YAML. La salida de texto plano está seleccionada por defecto, añada la bandera `--generate-config` o `-g` para seleccionar el modo de generación de YAML.

```bash
~ ▴ grep ^asn /etc/pathvector.yml
asn: 34553
~ ▴ pathvector match 13335
  Cloudflare SIX Seattle:
    asn: 13335
    neighbors:
      - 206.81.81.10
      - 2001:504:16::3417

  Cloudflare NWAX:
    asn: 13335
    neighbors:
      - 198.32.195.95
      - 2620:124:2000::95

  Cloudflare KCIX:
    asn: 13335
    neighbors:
      - 206.51.7.34
      - 2001:504:1b:1::34

  Cloudflare Speed-IX:
    asn: 13335
    neighbors:
      - 185.1.95.191
      - 2001:7f8:b7::a501:3335:1
```

## Texto sin formato

```bash
~ ▴ grep ^asn /etc/pathvector.yml
asn: 34553
~ ▴ pathvector match 13335
SIX Seattle: MTU 1500
  AS34553
  206.81.80.97
  2001:504:16::86f9
  
  AS13335
  206.81.81.10
  2001:504:16::3417

NWAX: Primary Peering VLAN
  AS34553
  198.32.195.28
  2620:124:2000::28
  
  AS13335
  198.32.195.95
  2620:124:2000::95

KCIX
  AS34553
  206.51.7.39
  2001:504:1b:1::39
  
  AS13335
  206.51.7.34
  2001:504:1b:1::34

Speed-IX: SPEED-IX
  AS34553
  185.1.95.166
  2001:7f8:b7::a503:4553:1
  
  AS13335
  185.1.95.191
  2001:7f8:b7::a501:3335:1
```

## ASN Arbitrario

El primer ASN se leerá del archivo de configuración `pathvector.yml` por defecto, pero puede suministrar un ASN alternativo con la bandera `--local-asn` o `-l` si quiere encontrar IXPs comunes para dos redes arbitrarias.

```bash
~ ▴ pathvector match -l 44977 34553
ARIX: Primary
  AS44977
  44.190.42.2
  2602:801:30ff::2
  
  AS34553
  44.190.42.3
  2602:801:30ff::3
```
