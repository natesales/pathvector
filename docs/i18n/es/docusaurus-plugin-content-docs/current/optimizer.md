---
sidebar_position: 5
---

# Optimización de rutas

Pathvector puede utilizar métricas de latencia y pérdida de paquetes para tomar decisiones de enrutamiento. El optimizador funciona enviando ping ICMP o UDP a diferentes redes pares y modificando el prefijo local de BGP según los umbrales de latencia media y pérdida de paquetes.

## Scripts de Alerta

Para ser notificado de un evento de optimización, puede añadir un script de alerta personalizado que Pathvector llamará cuando la latencia o la pérdida de paquetes alcancen o superen los umbrales configurados.
