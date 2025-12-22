# SensorFlow

# draft note

| Catégorie de capteur             | Intervalle courant              | Commentaire                                              |
| -------------------------------- | ------------------------------- | -------------------------------------------------------- |
| **Température / Humidité**       | 1 min – 15 min                  | Peu critique, on peut interpoler.                        |
| **Pression atmosphérique**       | 5 min – 30 min                  | Variations lentes, souvent 10–15 min.                    |
| **Luminosité / UV**              | 30 s – 5 min                    | Pour suivi conso ou ajustement d’éclairage.              |
| **Qualité de l’air (CO₂, VOC)**  | 1 min – 5 min                   | Détection de seuils, mais pas d’ultra‑faible latence.    |
| **Accéléromètre / Gyroscope**    | 10 Hz – 100 Hz (0.01 s – 0.1 s) | Très haute fréquence pour détection de chocs/mouvements. |
| **Détecteur de mouvement (PIR)** | Événementiel (push)             | Ne publie que quand mouvement détecté.                   |
| **GPS / Position**               | 1 s – 60 s                      | Suivi précis ou suivi “lent” selon l’usage.              |
| **Niveau de batterie**           | 5 min – 1 h                     | Statut d’état, pollué trop souvent gaspille.             |
| **Vibration / Son**              | 1 kHz – 10 kHz (streaming)      | Pour analyse en temps‑réel (p. ex. maintenance).         |
| **Compteur d’impulsions**        | Événementiel ou 1 min           | Comptage (eau, gaz), souvent push ou cron.               |


| Catégorie                               | Fields (mesures)                                                      | Unités / Type                          | Tags conseillés                                                            |
| --------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------- | --------------------------------------------------------------------------- |
| **Accéléromètre / Gyroscope**           | accel_x, accel_y, accel_z, gyro_x, gyro_y, gyro_z, temperature*        | g, °/s, °C*                             | device_id, location, sensor_id, firmware_version                            |
| **Détecteur de mouvement / Impulsions** | motion_detected, pulse_count, pulse_rate*                              | bool/int (0-1), int, Hz*                | device_id, sensor_id, type (motion/pulse), location                         |
| **Vibration / Son**                     | vib_rms, vib_peak, vib_freq, sound_db, sound_freq, sound_peak*         | m/s², Hz, dB SPL*                       | device_id, sensor_id, mounting_point                                        |
| **GPS / Position**                      | latitude, longitude, altitude, speed, heading, accuracy*               | °, m, m/s ou km/h*                      | device_id, source (GPS/GLONASS/etc.), mode (fix/float), firmware_version    |

Voici une version **simplifiée et condensée** de ton fichier `.md`, en mettant l’accent sur l’essentiel et en intégrant la notion de **TTL différent et cold storage via volumes** :

# Architecture IoT Multi-tenant

## 1️⃣ Types de tables et stratégie

### 🔹 Mutualisé standard
- 1 table par sensor standard partagé (ex: Accéléromètre, Gyroscope)
- Schéma fixe, ingestion massive
- TTL et partitions définis par l’opérateur
- Faible coût pour le client
- Pas de TTL ou index personnalisés

---

### 🔹 Custom mutualisé
- 1 table unique pour tous les sensors custom légers
- Schéma dynamique :
  ```sql
  attributes Map(String, String)
  metrics Map(String, Float64)


* Ingestion multi-tenant
* TTL standard, index minimal
* Faible à moyen volume
* Faible coût pour le client

---

### 🔹 Custom total / dédié

* Table dédiée par sensor ou type, optimisée pour performance
* Promotion automatique depuis custom mutualisé si :

  1. Volume élevé (>5k–10k events/sec ou >10% du trafic tenant)
  2. TTL personnalisé
  3. Schéma stable/fixe
  4. Index spécifique requis
  5. Requêtes fréquentes ou lourdes
  6. Partitionnement ou stockage spécial
  7. Large event (100+ metrics)
* Facturation adaptée au volume et aux fonctionnalités

---

## 2️⃣ Gestion des TTL et archives

* TTL sert à supprimer ou déplacer les données automatiquement

* On peut proposer plusieurs options TTL sur **tables mutualisées ou custom** :

  * 30j / 90j / 180j
  * Option **cold storage** via volumes séparés

    ```sql
    TTL timestamp + INTERVAL 180 DAY TO VOLUME 'cold';
    ```
  * Permet d’archiver les données pour réduire coût du hot storage

* TTL personnalisé **force la création d’une table dédiée**

---

## 3️⃣ Schéma type table custom mutualisé

```sql
CREATE TABLE sensor_custom (
    tenant_id UUID,
    sensor_id UUID,
    timestamp DateTime64(3),
    attributes Map(String, String),
    metrics Map(String, Float64),
    PRIMARY KEY (tenant_id, timestamp)
) ENGINE = MergeTree
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, sensor_id, timestamp);
```

* Supporte ingestion massive
* SQL possible sur Maps
* Index limité aux clés critiques pour éviter explosion de cardinalité

---

## 4️⃣ Paramètres clients (exposables et impact prix)

| Paramètre                | Description                   | Impact prix                                   |
| ------------------------ | ----------------------------- | --------------------------------------------- |
| TTL                      | Durée de rétention            | Custom TTL → table dédiée obligatoire         |
| Volume / rows/sec        | Limite de données / débit     | Débit élevé → table dédiée possible           |
| Index personnalisés      | Index sur metrics/attributes  | CPU + stockage → option payante               |
| Partitionnement          | Mois / semaine / jour         | Custom → table dédiée                         |
| Table dédiée (promotion) | Sensor actif ou plan premium  | Isolé et optimisé → facturable                |
| Support schema           | Metrics fixes ou Map flexible | Map flexible → mutualisé, fixe → table dédiée |
| Accès rapide / SLA       | Query temps réel vs batch     | Table dédiée si SLA stricte nécessaire        |

---

## 5️⃣ Décision mutualisé vs custom

| Critère sensor custom                   | Résultat         |
| --------------------------------------- | ---------------- |
| Volume faible / moyen                   | Custom mutualisé |
| Map légère (3–10 metrics)               | Custom mutualisé |
| TTL standard, index minimal             | Custom mutualisé |
| Volume élevé, SLA strict, TTL custom    | Custom dédié     |
| Schéma fixe / index spécifique          | Custom dédié     |
| Requêtes lourdes / dashboards fréquents | Custom dédié     |
| Large event / Maps larges               | Custom dédié     |

---

## 6️⃣ Principes clés

* **Mutualisé** : sensors standard → simple, low-cost
* **Custom mutualisé** : sensors légers → flexibilité, low-cost
* **Custom dédié** : sensors volumineux ou TTL/index spécifiques → performance + coût adapté
* **Promotion automatique** : custom mutualisé → dédié selon volume, TTL, index ou SLA
* **TTL et cold storage** : plusieurs tables TTL possibles, option cold storage via volumes, TTL personnalisé = table dédiée
* ClickHouse gère ingestion massive via `MergeTree`, `PARTITION BY`, `ORDER BY`
* SQL possible sur Maps, index limité pour éviter explosion cardinalité



| Champ       | Unité envoyée | Unité stockée | Conversion à l’ingestion |
| ----------- | ------------- | ------------- | ------------------------ |
| temperature | °F            | °C            | °C = (°F - 32) * 5/9     |
| accel_x/y/z | g             | m/s²          | m/s² = g * 9.80665       |
| gyro_x/y/z  | rad/s         | °/s           | °/s = rad/s * 180/π      |
