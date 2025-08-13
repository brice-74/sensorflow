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


ClickHouse (cold storage)
https://github.com/ClickHouse/clickhouse-go
https://github.com/InfluxCommunity/influxdb3-go
https://github.com/segmentio/kafka-go

| Catégorie                               | Fields (mesures)                                                      | Unités / Type                          | Tags conseillés                                                            |
| --------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------- | --------------------------------------------------------------------------- |
| **Accéléromètre / Gyroscope**           | accel_x, accel_y, accel_z, gyro_x, gyro_y, gyro_z, temperature*        | g, °/s, °C*                             | device_id, location, sensor_id, firmware_version                            |
| **Détecteur de mouvement / Impulsions** | motion_detected, pulse_count, pulse_rate*                              | bool/int (0-1), int, Hz*                | device_id, sensor_id, type (motion/pulse), location                         |
| **Vibration / Son**                     | vib_rms, vib_peak, vib_freq, sound_db, sound_freq, sound_peak*         | m/s², Hz, dB SPL*                       | device_id, sensor_id, mounting_point                                        |
| **GPS / Position**                      | latitude, longitude, altitude, speed, heading, accuracy*               | °, m, m/s ou km/h*                      | device_id, source (GPS/GLONASS/etc.), mode (fix/float), firmware_version    |
