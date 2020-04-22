# Auto Rest IoT Service #

## REST Interface

Dieser kleine Service ermöglicht es, schnell eine permanente Datenspeicherung über ein einfaches REST/gRPC Interface zu ermöglichen. 

Der Service definiert einen eigenen API Key. Dieser Key muss bei jeder REST Kommunikation als Header X-mcs-apikey mit gesendet werden. Bei einer Installation kann der Apikey aus der Console ausgelesen werden.

Und jeder Call muss authentifiziert werden. Dazu werden bei einer Neuinstallation direkt mehrere User mit unterschiedlichen Rollen angelegt. Die Defaults findet man hier unter **User**. 

Ein Service kann mehrere Backends verwalten. 

## Konfiguration des Service

```yaml
# port of the http server
port: 8080 
# port of the https server
sslport: 8443
# this is the servicURL from outside
serviceURL: http://127.0.0.1:8080
# this is the registry URL from inside this service for consul as service registry
registryURL: 
# this is the system id of this service. services in a cluster mode should have the same system id.
systemID: autorest-srv
#sercret file for storing usernames and passwords
secretfile: /tmp/storage/config/secret.yaml
#where the configuration files of the backends are
backendpath: configs/backends
#allow data saving without a registered backend
allowAnonymousBackend: true

# configuration of the gelf logging server
logging:
    gelf-url: 
    gelf-port: 

# healthcheck configuration
healthcheck:
    # automatically check the health of this service every ## seconds
    period: 30
# configuration of the mongo storage
mongodb:
    host: 127.0.0.1 #mongo host ip (comma seperated ip's with a cluster installation )
    port: 27017 # mongo host port
    username: #username for the database (should be at last dbadmin, and read/write access)
    password: #password for the user
    authdb: backend1 # database to authenticate against
    database: backend1 # database for the data

```



## Storage

Als Storage wird derzeit nur MongoDB unterstützt. 

Bei der Mongo Storage Implementierung werden die verschiedenen Backends allerdings in einer Datenbank abgelegt. Einzelne Modelle werden in jeweils einer Collection abgelegt.  Der Collectionname besteht aus dem Backendnamen  "." und dem Modellnamen. 

### Hint

Um eine neue Mongodatenbank anzulegen, müssen folgende Kommandos auf der Mongo Console ausgeführt werden:

```json
#create a new db named backend1
use backend1
#create a db admin on db backend1 with user backend1 with password backend1
db.createUser({ user: "backend1", pwd: "backend1", roles: [ "readWrite", "dbAdmin", { role: "dbOwner", db: "backend1" } ]})

```



## Anlegen eines Backends mittels yaml Datei

Zum Erzeugen eines neuen Backendes benötigt man eine eigene yaml Datei. (Später wird dieses auch über das REST Interface selber funktionieren). Jedes Backend hat einen eindeutigen Namen.

Der Backendname darf max. 60 Zeichen lang sein und sollte nur aus Kleinbuchstaben bestehen. Sonderzeichen wie "$", ".", "_", oder "-" sind nicht erlaubt. Auch ein Leerzeichen " " darf nicht verwendet werden. 

Jedes Backend besteht nun aus einer Reihe von Modellen. Ein Modell kann man sich als eine Tabelle vorstellen. Will man Daten in eine Tabelle ablegen, muss man ein Modell dafür definieren.
Jedes Modell hat einen eigenen Namen und definiert eine Reihe von Feldern/Attributen. Grundsätzlich werden alle übergebenden Attribute gespeichert, auch wenn Sie hier nicht definiert wurden. Die Definition dient einerseits der besseren Indexierung. D.h. will man einen Suchindex für ein Attribute oder eine Kombination mehrerer Attribute anlegen, müssen die verwendeten Attribute hier zumindest mit Type definiert werden. 
Auch eine Attributvalidierung (wie z.B. auch das Mandatory) erfordert die Definition des jeweiligen Attributes hier.

Typische JSON Attribute/Objektverschachtelungen sind grundsätzlich erlaubt. Neben den Attributen kann man pro Modell auch noch eine Reihe von Suchindizies definieren, um einen schnelleren Zugriff zu ermöglichen. Eine Besonderheit stellt der Volltextindex dar. Man kann pro Modell **einen** Volltextindex definieren. Dabei wird dann jedes angegebene Feld in diesem Index gespeichert und über eine eingängige Suchsyntax wieder findbar abgelegt. Dazu mehr im Kapitel Suche.

```yaml
applicationname: schematicworld  #name without whitespaces and special charaters
description: Willies World Schematics Database #description of the backend
models:  #definition of the different models
    - name: schematics #name of the models, no whitespaces or special chars
      description: This are the different schematics # a model description
      fields: #definition of the fields/attributes
        - name: manufacturer #name of the field, , no whitespaces or special chars
          type: string  #string, int, float, bool, map, file, more to come...
          mandatory: true #internal validator for present
          collection: false #field is a collection of types 
        - name: model
          type: string
          mandatory: true
          collection: false
        - name: tags
          type: string
          mandatory: false
          collection: true
      indexes:
        - name: fulltext #revered name for the fulltext index
          fields: # defining which fields should be in that index
            - manufacturer
            - model
            - tags
        - name: manufacturer #single field index
          fields:
            - manufacturer
        - name: tags #single field index on a collection field 
          fields:
            - tags

```

## Dateien

Dateien können pro Backend in das reserviert Model files abgelegt werden. Sollen diese einem Modell zugeordnet werden, sollte man ein Attribut vom Typ ID anlegen. Der Service stellt dann automatisch die Referenzierung sicher. D.h. wird eine Modelinstanz aus den Modellen gelöscht, wird automatisch die referenzierte Instanz mit gelöscht. (Eine Referenzzählung wird nicht vorgenommen, d.h. wird ein und dieselbe Dateiinstanz in verschiedenen Modellen verwendet, und eines der Modelle gelöscht, wird die Dateiinstanz mit gelöscht.) Dieses Verhalten kann mit dem Header `X-mcs-deleteref: false` verhindert werden.

## User

Folgende User mit folgenden Rollen werden automatisch angelegt:

- Admin, pwd: admin, roles: admin
- Editor, pwd: editor, roles: edit
- guest, pwd: guest, roles: read

## Indizes

Zur schnelleren Suche definiert das System automatisch diverse Indizes. Zu jedem definierten Attribut wird automatisch ein Index erstellt. Und zusätzlich wird noch ein spezieller Volltextindex über alle definierten Attribute erzeugt. Durch eine eigene Definition von einem Index mit dem gleichen Namen können die automatisch erzeugten Indizes überschrieben werden. Bitte denken Sie daran, dass eine Änderung eines Index nicht möglich ist. Soll ein Index geändert werden, muss dieser vorher aus dem System gelöscht werden. Entweder per API oder aber direkt auf der Datenbank. Beim nächsten Neustart des Service oder bei einem Refesh über das Admin API wird dann der neue Index erzeugt. 

Ein Index kann als `unique` gekennzeichnet werden. Das Attribute (bzw. die Kombination der Attribute) muss dann eindeutig sein, d.h. kein Wert darf doppelt vorkommen. 
Beispiel:

```yaml
...
   indexes:
   ...
      - name: foreignid
        unique: true
        fields:
          - foreignid
...
```



## Datasource

Pro Applikation kann man eine Liste von Datasources angeben. Der Service wird dann automatisch über diese Importkanäle Daten abholen und in das System importieren. 

gemeinsame Properties:

```yaml
...
datasources:
  - name: temp_wohnzimmer
    type: mqtt
    destination: temperatur
    config: 
```

**name**: eindeutiger Name der Quelldefinition

**type**: Typ des Importplugin to use. Derzeit unterstützt: MQTT

**destination**: Welches Model soll als Speicher dienen

**config**: pluginspezifische Konfigurationseinstellungen

### MQTT

```yaml
...
datasources:
  - name: temp_wohnzimmer
    type: mqtt
    destination: temperatur
    config: 
      broker: 127.0.0.1:1883
      topic: stat/temperatur/wohnzimmer
      payload: application/json
      username: temp
      password: temp
      addTopicAsAttribute: topic
  - name: temp_kueche
    type: mqtt
    destination: temperatur
    config: 
      broker: 127.0.0.1:1883
      topic: stat/temperatur/kueche
      payload: application/json
      username: temp
      password: temp
      addTopicAsAttribute: topic
...
```

**broker**: IP und Port  des MQTT Brokers

**topic**:  Topic aus welchem die Daten importiert werden sollen

**payload**: Mimetyp der payload auf dem Topic. Derzeit unterstützt: 

- JSON: die Payload enthält ein Json Objekt. Dieses wird dann automatisch auf die Model Struktur  gemapped. 

**username, password**: Authentifizierung gegen den Broker

**addTopicAsAttribute**: das Topic wird zusätzlich in ein Attribute mit dem definierten Namen gespeichert.