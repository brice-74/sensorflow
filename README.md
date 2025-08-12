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