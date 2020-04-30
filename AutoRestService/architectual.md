# Auto Rest IoT Service #

## Einleitung

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

Die Secrets zu diesem Service, im speziellen username und Passwort zur MongoDB werden in einer speziellen Datei secrets.yaml abgelegt. Diese wird über den Configeintrag `secretfile` bestimmt.

Inhalt:

```yaml
mongodb:
    username: backend1
    password: backend1
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
backendname: schematicworld  #name without whitespaces and special charaters
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
        - name: $fulltext #reserved name for the fulltext index
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
        - name: Foreignid #single unique field index on a collection field 
          unique: true
          fields:
            - Foreignid
```

## Dateien

Dateien können pro Backend in das reserviert Model files abgelegt werden. Sollen diese einem Modell zugeordnet werden, sollte man ein Attribut vom Typ `file` anlegen. Der Service stellt dann automatisch die Referenzierung sicher. D.h. wird eine Modelinstanz aus den Modellen gelöscht, wird automatisch die referenzierte Instanz mit gelöscht. (Eine Referenzzählung wird nicht vorgenommen, d.h. wird ein und dieselbe Dateiinstanz in verschiedenen Modellen verwendet, und eines der Modelle gelöscht, wird die Dateiinstanz mit gelöscht.) Dieses Verhalten kann mit dem Header `X-mcs-deleteref: false` verhindert werden.

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
    destination: 
      - $model.temperatur
    config: 
```

**name**: eindeutiger Name der Quelldefinition

**type**: Typ des Importplugin to use. Derzeit unterstützt: MQTT

**destination**: Welches Destination Plugin soll zum speichern verwendet werden. Zur Speicherung in der internen Datenbank wird dem Modelnamen ein `$model`. vorangestellt.

Beispiel: das aufbereitete JSON soll in sowohl das interne Modell `temperature` abgelegt werden als auch  per MQTT auf dem einem Topic bereitgestellt werden.

```yaml
  - name: temp_kueche
    type: mqtt
    destinations: 
      - $model.temperatur 
      - mqtt_sensors_temperatur
    rule: tasmota_ds18b20
    config: 
      broker: 192.168.178.12:1883
      topic: tele/tasmota_63E6F8/SENSOR
      qos: 0
      payload: application/json
      username: temp
      password: temp
      addTopicAsAttribute: topic
```

**rule**: definiert welche Regel vor der Weiterverarbeitung der Daten ausgeführt werden sollen

**config**: enthält dabei die pluginspezifische Konfigurationseinstellungen

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
      qos: 0
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
      qos: 0
      payload: application/json
      username: temp
      password: temp
      addTopicAsAttribute: topic
  - name: temp_simple_time
    type: mqtt
    destination: temperatur
    config: 
      broker: 127.0.0.1:1883
      topic: stat/temperatur/simple/time
      qos: 0
      payload: application/x.simple
      username: temp
      password: temp
      addTopicAsAttribute: topic
      simpleValueAttribute: vtime
...
```

**broker**: IP und Port  des MQTT Brokers

**topic**:  Topic aus welchem die Daten importiert werden sollen

**qos**: Qualtity of Service, 0 ,1 oder 2. Als Standard wird immer 0 angenommen.

**payload**: beschreibt den Mimetypen der Payload auf dem Topic. Derzeit unterstützt: 

- **application/json**: die Payload enthält ein Json Objekt. Dieses wird dann automatisch auf die Model Struktur  gemapped. 
- **application/x.simple**: die Payload besteht aus nur einem einzigen Wert. Dieser Wert wird dann in das Attribut **simpleValueAttribute** abgelegt. ist dieses Attribut in der Attributliste des Models definiert findet eine automatische Typkonvertierung statt. Für Zeitattribute können 2 verschiedene Format verwendet werden. Einerseits eine normale Ganzzahl, die die Millisekunden seit 1.1.1970 angibt (UNIX Zeitstempel), andererseits kann der Zeitstempel auch im [RFC3339 Format](https://en.wikipedia.org/wiki/ISO_8601)  als String gesendet werden, z.B.: 2020-04-23T06:20:16.730+00:00

**username, password**: Authentifizierung gegen den Broker

**addTopicAsAttribute**: das Topic wird zusätzlich in ein Attribute mit dem definierten Namen gespeichert.

**simpleValueAttribute**: kommt der zu speichernde Wert als einfacher Wert, wird das hier benannte Attribut zur Ablage benutzt.

## Transformation Rules

Nicht immer sollen die aus einer Datasource gelesenen Daten ohne Modifikationen in den Storage geschrieben werden. Für den Fall einer JSON Payload können Rules definiert werden, mit denen man das JSON Objekt transformieren kann, bevor es in den Storage gespeichert wird. Für die Transformation wird die Go Bibliothek [Kazaam](https://github.com/qntfy/kazaam) verwendet. Die Bibliothek verwendet als Definitionssprache JSON. Da die AutoRestIoT Konfiguration in Dateien aber immer in YAML vorliegt,  hier die Definition der verschiedenen Transformationen in YAML übersetzt nach `yaml`. 

```yaml
datasources:
  - name: temp_wohnzimmer
    type: mqtt
    destination: temperatur
	rule: tasmota_ds18b20
    config: 
      broker: 127.0.0.1:1883
      topic: stat/temperatur/wohnzimmer
      payload: application/json
      username: temp
      password: temp
      addTopicAsAttribute: topic
...
rules:
  - name: tasmota_ds18b20
    description: transforming the tasmota json structure of the DS18B20 into my simple structure
	transform: 
	  - operation: shift
	    spec: 
		  Temperatur: DS18B20.Temperatur
	  - operation: delete
	    spec: 
		  path: DS18B20

```

In der Definition der Data wird durch **rule** der Name einer anzuwendenden Regel angegeben.  Die Regel selber werden im Bereich `rules` definiert.

**name**: definiert den Namen der Regel. Innerhalb einer Anwendung müssen diese Namen eindeutig sein. Vordefinierte Regeln sind noch nicht imülementiert

**description**: gibt eine kurze Beschreibung der Regel

**transform**: defniert nun die verschiedenen Transformationsregeln.

## Transformationsregeln
Derzeit werden folgende Transformation unterstützt:
- shift
- concat
- coalesce
- extract
- timestamp
- uuid
- default
- pass
- delete

### Shift
Die Shift-Transformation wird zum Neuzuordnen von Feldern verwendet.
Die Spezifikation unterstützt jsonpath-ähnliche JSON-Zugriffe. Konkret

```yaml
- operation: shift
  spec:
    object.id: doc.uid
    gid2: doc.guid[1]
    allGuids: doc.guidObjects[*].id
```

```javascript
{
  "operation": "shift",
  "spec": {
    "object.id": "doc.uid",
    "gid2": "doc.guid[1]",
    "allGuids": "doc.guidObjects[*].id"
  }
}
```

JSON-Nachricht

```javascript
{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  },
  "top-level-key": null
}
```

wird zu 
```javascript
{
  "object": {
    "id": 12345
  },
  "gid2": "guid2",
  "allGuids": ["guid0", "guid2", "guid4"]
}
```

Die Implementierung von jsonpath unterstützt einige Sonderfälle:

- * **Array-Zugriffe** : Ruft das n-te Element aus dem Array ab
- * **Array-Platzhalter** : Durch Indizieren eines Arrays mit `[*]` wird jedes übereinstimmende Element in einem Array zurückgegeben
- * **Objekterfassung der obersten Ebene**: Durch Zuordnen von "$" in ein Feld wird das gesamte Originalobjekt unter dem angeforderten Schlüssel verschachtelt
- * **Array anhängen / voranstellen und setzen** : Ein Array mit `[+]` und `[-]` anhängen und voranstellen. Der Versuch, ein nicht vorhandenes Array-Element zu schreiben, führt nach Bedarf zu einem Null-Padding, um dieses Element am angegebenen Index hinzuzufügen (nützlich bei "inplace").

Die Shift-Transformation unterstützt auch ein `require` -Feld. Wenn auf `true` gesetzt,
wird ein Fehler erzeugt, wenn *einer* der Pfade im Quell-JSON nicht vorhanden ist.

### Concat
Die Concat-Transformation ermöglicht die Kombination von Feldern und Literalzeichenfolgen zu einem einzigen Zeichenfolgenwert.

```yaml
- operation: concat
  spec:
    sources:
    - value: TEST
    - path: a.timestamp
    targetPath: a.timestamp
    delim: ","
```

```javascript
{
    "operation": "concat",
    "spec": {
        "sources": [{
            "value": "TEST"
        }, {
            "path": "a.timestamp"
        }],
        "targetPath": "a.timestamp",
        "delim": ","
    }
}
```

JSON-Nachricht
```javascript
{
    "a": {
        "timestamp": 1481305274
    }
}
```

wird zu
```javascript
{
    "a": {
        "timestamp": "TEST,1481305274"
    }
}
```

Anmerkungen:

- **Quellen**: Liste der zu kombinierenden Elemente (in der angegebenen Reihenfolge)
   - Literalwerte werden über `value` angegeben
   - Feldwerte werden über `path` angegeben (unterstützt die gleiche Adressierung wie `shift`)
- **targetPath**: Wo soll die resultierende Zeichenfolge platziert werden?
  
   - Wenn dies ein vorhandener Pfad ist, ersetzt das Ergebnis den aktuellen Wert.
- **delim**: Optionales Trennzeichen

Die Concat-Transformation unterstützt auch ein `require`-Feld. Wenn auf `true` gesetzt,
wird ein Fehler erzeugt, wenn *einer* der Pfade im Quell-JSON nicht vorhanden ist.

### Coalesce
Eine Coalesce-Transformation bietet die Möglichkeit, mehrere mögliche Schlüssel zu überprüfen, um einen gewünschten Wert zu finden. Der erste gefundene passende Schlüssel wird zurückgegeben.

```yaml
- operation: coalesce
  spec:
    firstObjectId:
    - doc.guidObjects[0].uid
    - doc.guidObjects[0].id
```



```javascript
{
  "operation": "coalesce",
  "spec": {
    "firstObjectId": ["doc.guidObjects[0].uid", "doc.guidObjects[0].id"]
  }
}
```

JSON-Nachricht
```javascript
{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  }
}
```

wird zu
```javascript
{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  },
  "firstObjectId": "guid0"
}
```

Coalesce unterstützt auch ein `ignore`-Array in der Spezifikation. Wenn ein ansonsten übereinstimmender Schlüssel den Wert "Ignorieren" hat, wird er nicht als Übereinstimmung betrachtet.
Dies ist z.B. für leere Zeichenketten interessant

```javascript
{
  "operation": "coalesce",
  "spec": {
    "ignore": [""],
    "firstObjectId": ["doc.guidObjects[0].uid", "doc.guidObjects[0].id"]
  }
}
```

### Extract
Die `extract`-Transform bietet die Möglichkeit, ein Unterobjekt auszuwählen und dieses Unterobjekt als Objekt der obersten Ebene zurückgeben zu lassen.

Beispiel

```yaml
- operation: extract
  spec:
    path: doc.guidObjects[0].path.to.subobject
```

```javascript
{
  "operation": "extract",
  "spec": {
    "path": "doc.guidObjects[0].path.to.subobject"
  }
}
```

JSON-Nachricht
```json
{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"path": {"to": {"subobject": {"name": "the.subobject", "field", "field.in.subobject"}}}}, {"id": "guid2"}, {"id": "guid4"}]
  }
}
```

wird zu
```javascript
{
  "name": "the.subobject",
  "field": "field.in.subobject"
}
```

### Timestamp
Ein `timestamp`-Transformation transformiert und formatiert Zeitzeichenfolgen in der Golang
Syntax. Diese Transformation unterstützt den Operator `$now` für `inputFormat`, der den aktuellen Zeitstempel gemäß dem `outputFormat` formatiert und am angegebenen Pfad einträgt.
`$unix` wird sowohl für Eingabe- als auch für Ausgabeformate als Unix-Zeit unterstützt
Anzahl der Sekunden seit dem 1. Januar 1970 UTC als Ganzzahl. `$unixext` ist eine Abwandlung für Unixdaten in Millisekunden seit Epoche. Diese wird gerne in IoT Umgebungen verwendet.

```yaml
- operation: timestamp
  spec: 
    timestamp[0]:
      inputFormat: Mon Jan _2 15:04:05 -0700 2006
      outputFormat: '2006-01-02T15:04:05-0700'
    nowTimestamp:
      inputFormat: "$now"
      outputFormat: '2006-01-02T15:04:05-0700'
    epochTimestamp:
      inputFormat: '2006-01-02T15:04:05-0700'
      outputFormat: "$unix"
```

```javascript
{
  "operation": "timestamp",
  "spec": {
    "timestamp[0]": {
      "inputFormat": "Mon Jan _2 15:04:05 -0700 2006",
      "outputFormat": "2006-01-02T15:04:05-0700"
    },
    "nowTimestamp": {
      "inputFormat": "$now",
      "outputFormat": "2006-01-02T15:04:05-0700"
    },
    "epochTimestamp": {
      "inputFormat": "2006-01-02T15:04:05-0700",
      "outputFormat": "$unix"
    }
  }
}
```

JSON-Nachricht

```javascript
{
  "timestamp": [
    "Sat Jul 22 08:15:27 +0000 2017",
    "Sun Jul 23 08:15:27 +0000 2017",
    "Mon Jul 24 08:15:27 +0000 2017"
  ]
}
```

wird zu
```javascript
{
  "timestamp": [
    "2017-07-22T08:15:27+0000",
    "Sun Jul 23 08:15:27 +0000 2017",
    "Mon Jul 24 08:15:27 +0000 2017"
  ]
  "nowTimestamp": "2017-09-08T19:15:27+0000"
}
```

### UUID
Eine UUID-Transformation generiert eine UUID basierend auf den Spezifikationen UUIDv3, UUIDv4, UUIDv5.

Für UUIDv4 ist die Spezifikation einfach

```yaml
- operation: uuid
  spec:
    doc.uuid:
      version: 4
```

```javascript
{
    "operation": "uuid",
    "spec": {
        "doc.uuid": {
            "version": 4, //required
        }
    }
}
```

JSON-Nachricht
```javascript
{
  "doc": {
    "author_id": 11122112,
    "document_id": 223323,
    "meta": {
      "id": 23
    }
  }
}
```

wird zu
```javascript
{
  "doc": {
    "author_id": 11122112,
    "document_id": 223323,
    "meta": {
      "id": 23
    }
    "uuid": "f03bacc1-f4e0-4371-a5c5-e8160d3d6c0c"
  }
}
```

Für UUIDv3 & UUIDV5 sind die Konfigurationen etwas komplexer. Diese erfordern einen Namensraum, der bereits eine gültige UUID ist, und eine Reihe von Pfaden, die UUIDs basierend auf dem Wert dieses Pfads generieren. Wenn dieser Pfad im eingehenden Dokument nicht vorhanden ist, wird stattdessen ein Standardfeld verwendet. 
**Hinweis** Diese beiden Felder müssen Zeichenfolgen sein.
**Zusätzlich** können Sie die 4 vordefinierten Namespaces wie "DNS", "URL", "OID" und "X500" im Feld "Namensraum" verwenden, andernfalls übergeben Sie Ihre eigene UUID.

```yaml
- operation: uuid
  spec:
    doc.uuid:
      version: 5
      namespace: DNS
      names:
      - path: doc.author_name
        default: some string
      - path: doc.type
        default: another string
```

```javascript
{
   "operation":"uuid",
   "spec":{
      "doc.uuid":{
         "version":5,
         "namespace":"DNS",
         "names":[
            {"path":"doc.author_name", "default":"some string"},
            {"path":"doc.type", "default":"another string"}
         ]
      }
   }
}
```

JSON-Nachricht
```javascript
{
  "doc": {
    "author_name": "jason",
    "type": "secret-document"
    "document_id": 223323,
    "meta": {
      "id": 23
    }
  }
}
```

wird zu
```javascript
{
  "doc": {
    "author_name": "jason",
    "type": "secret-document",
    "document_id": 223323,
    "meta": {
      "id": 23
    },
    "uuid": "f03bacc1-f4e0-4371-a7c5-e8160d3d6c0c"
  }
}
```


### Default
Eine Default-Transformation bietet die Möglichkeit, den Wert eines Schlüssels explizit festzulegen. Zum Beispiel

```yaml
- operation: default
  spec:
    type: message
```



```javascript
{
  "operation": "default",
  "spec": {
    "type": "message"
  }
}
```
würde sicherstellen, dass das Ausgabe-JSON `{"type":"message"}` enthält.


### Delete
Eine Delete-Transformation bietet die Möglichkeit, vorhandene Schlüssel zu löschen.

```yaml
- operation: delete
  spec:
    paths:
    - doc.uid
    - doc.guidObjects[1]
```



```javascript
{
  "operation": "delete",
  "spec": {
    "paths": ["doc.uid", "doc.guidObjects[1]"]
  }
}
```

JSON-Nachricht
```javascript
{
  "doc": {
    "uid": 12345,
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid2"}, {"id": "guid4"}]
  }
}
```

wird zu
```javascript
{
  "doc": {
    "guid": ["guid0", "guid2", "guid4"],
    "guidObjects": [{"id": "guid0"}, {"id": "guid4"}]
  }
}
```

## Destinations

Neben der internen Datensenke Datenbank, können noch weitere Empfänger definiert werden. 

### MQTT

Um die Message auf einem MQTT Server zu veröffentlichen, sind folgende Einstellungen nötig. 

```yaml
destinations:
  - name: mqtt_sensors_temperatur
    type: mqtt
    config: 
      broker: 192.168.178.12:1883
      topic: stat/temperatur
      qos: 0
      payload: application/json
      username: temp
      password: temp
```

Die Beschreibung der Parameter entnehmen Sie bitte aus dem Kapitel Datasource/MQTT



## REST Interface

Hier nun folgt die Beschreibung des REST Interfaces. Beispielhafte REST Calls sind als Postman Collection im Ordner test/postman vorhanden. 

Beispielhaft werden hier alle Calls als Calls auf den lokalen Server (127.0.0.1) mit dem Port 9443 bereitgestellt. Bei einer anderen Serverinstanz bitte entsprechend ändern.

### Admin API

Security: Ja, Authentifizierung derzeit als BasicAuth.
Role: admin

zus. Header:

**X-mcs-system**: autorest-srv  (Konfigurationseinstellung: systemID)

**X-mcs-apikey**: {uuid} 
Wird beim Starten des Servers auf der Konsole ausgegeben. 

```
...
2020/04/29 08:43:04 systemid: autorest-srv
2020/04/29 08:43:04 apikey: 5854d123dd25f310395954f7c450171c
2020/04/29 08:43:04 ssl: true
...
```



#### Liste alle Backends

**Request**: **GET**: https://127.0.0.1:9443/api/v1/admin/backends

**Beschreibung**: Liefert eine Liste mit allen Backenddefinitioninformationen. D.h. pro Backend werden nur der Name, die Beschreibung und die URL auf die Definition ausgeliefert.

**Request**: **GET**: https://127.0.0.1:9443/api/v1/admin/backends

**Response**:

```JSON
[
    {
        "Name": "sensors",
        "Description": "sensor model für storing and retrieving sensor data",
        "URL": "https://127.0.0.1:9443/api/v1/admin/backends/sensors/"
    },
    {
        "Name": "mybe",
        "Description": "",
        "URL": "https://127.0.0.1:9443/api/v1/admin/backends/mybe/"
    }
]
```

#### Definition eines Backends

**GET**: https://127.0.0.1:9443/api/v1/admin/backends/{backendname}/

**Beschreibung**: Liefert die Definition eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/

**Response**:

```JSON
{
    "backendname": "sensors",
    "description": "sensor model für storing and retrieving sensor data",
    "models": [
        {
            "name": "temperatur",
            "description": "",
            "fields": [
                {
                    "name": "temperatur",
                    "type": "float",
                    "mandatory": false,
                    "collection": false
                },
                {
                    "name": "source",
                    "type": "string",
                    "mandatory": false,
                    "collection": false
                }
            ],
            "indexes": null
        }
    ],
    "datasources": [
        {
            "name": "temp_wohnzimmer",
            "type": "mqtt",
            "destination": "temperatur",
            "rule": "tasmota_ds18b20",
            "config": {
                "broker": "127.0.0.1:1883",
                "topic": "stat/temperatur/wohnzimmer",
                "payload": "application/json",
                "username": "temp",
                "password": "temp",
                "addTopicAsAttribute": "topic",
                "simpleValueAttribute": ""
            }
        },
        {
            "name": "temp_kueche",
            "type": "mqtt",
            "destination": "temperatur",
            "rule": "tasmota_ds18b20",
            "config": {
                "broker": "127.0.0.1:1883",
                "topic": "tele/tasmota_63E6F8/SENSOR",
                "payload": "application/json",
                "username": "temp",
                "password": "temp",
                "addTopicAsAttribute": "topic",
                "simpleValueAttribute": ""
            }
        }
    ],
    "rules": [
        {
            "name": "tasmota_ds18b20",
            "description": "transforming the tasmota json structure of the DS18B20 into my simple structure",
            "transform": [
                {
                    "operation": "shift",
                    "spec": {
                        "TempUnit": "TempUnit",
                        "Temperature": "DS18B20.Temperature"
                    }
                }
            ]
        },
        {
            "name": "hm_temp_simple",
            "description": "handle homematic temperatur rightly",
            "transform": [
                {
                    "operation": "shift",
                    "spec": {
                        "Datetime": "ts",
                        "Temperature": "val",
                        "Timestamp": "ts"
                    }
                },
                {
                    "operation": "default",
                    "spec": {
                        "TempUnit": "°C"
                    }
                },
                {
                    "operation": "timestamp",
                    "spec": {
                        "Datetime": {
                            "inputFormat": "$unixext",
                            "outputFormat": "2006-01-02T15:04:05-0700"
                        }
                    }
                }
            ]
        }
    ]
}
```

#### Neues Backends anlegen

**POST**: https://127.0.0.1:9443/api/v1/admin/backends/

**Beschreibung**: Liefert die Definition eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/

**Response**: **Not implemented Yet**

#### Daten eines Backends löschen

**DELETE**: https://127.0.0.1:9443/api/v1/admin/backends/

**Beschreibung**: Löscht alle Daten eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/dropdata

**Response**: 

```json
{
    "backend": "sensors",
    "msg": "backend sensors deleted. All data destroyed."
}
```

#### Liste aller Modelle eines Backends 

**GET**: https://127.0.0.1:9443/api/v1/admin/backends/{backendname}/models

**Beschreibung**: Liefert eine Liste aller Modelle eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/models

**Response**: 

```json
[
    {
        "name": "temperatur",
        "description": "",
        "fields": [
            {
                "name": "temperatur",
                "type": "float",
                "mandatory": false,
                "collection": false
            },
            {
                "name": "source",
                "type": "string",
                "mandatory": false,
                "collection": false
            }
        ],
        "indexes": null
    }
]
```

#### Liste aller Datenquellen eines Backends 

**GET**: https://127.0.0.1:9443/api/v1/admin/backends/{backendname}/datasources

**Beschreibung**: Liefert eine Liste aller Datenquellen eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/datasources

**Response**: 

```json
[
    {
        "name": "temp_wohnzimmer",
        "type": "mqtt",
        "destination": "temperatur",
        "rule": "tasmota_ds18b20",
        "config": {
            "broker": "127.0.0.1:1883",
            "topic": "stat/temperatur/wohnzimmer",
            "payload": "application/json",
            "username": "temp",
            "password": "temp",
            "addTopicAsAttribute": "topic",
            "simpleValueAttribute": ""
        }
    },
    {
        "name": "temp_kueche",
        "type": "mqtt",
        "destination": "temperatur",
        "rule": "tasmota_ds18b20",
        "config": {
            "broker": "127.0.0.1:1883",
            "topic": "tele/tasmota_63E6F8/SENSOR",
            "payload": "application/json",
            "username": "temp",
            "password": "temp",
            "addTopicAsAttribute": "topic",
            "simpleValueAttribute": ""
        }
    }
]
```

#### Liste aller Transformationsregeln eines Backends 

**GET**: https://127.0.0.1:9443/api/v1/admin/backends/{backendname}/rules

**Beschreibung**: Liefert eine Liste aller Transformationsregelen eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/rules

**Response**: 

```json
[
    {
        "name": "tasmota_ds18b20",
        "description": "transforming the tasmota json structure of the DS18B20 into my simple structure",
        "transform": [
            {
                "operation": "shift",
                "spec": {
                    "TempUnit": "TempUnit",
                    "Temperature": "DS18B20.Temperature"
                }
            }
        ]
    },
    {
        "name": "hm_temp_simple",
        "description": "handle homematic temperatur rightly",
        "transform": [
            {
                "operation": "shift",
                "spec": {
                    "Datetime": "ts",
                    "Temperature": "val",
                    "Timestamp": "ts"
                }
            },
            {
                "operation": "default",
                "spec": {
                    "TempUnit": "°C"
                }
            },
            {
                "operation": "timestamp",
                "spec": {
                    "Datetime": {
                        "inputFormat": "$unixext",
                        "outputFormat": "2006-01-02T15:04:05-0700"
                    }
                }
            }
        ]
    }
]
```

#### Definition einer Transformationsregel eines Backends 

**GET**: https://127.0.0.1:9443/api/v1/admin/backends/{backendname}/rules/{rulename}

**Beschreibung**: Definition einer Transformationsregel eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/rules/hm_temp_simple

**Response**: 

```json
{
    "name": "hm_temp_simple",
    "description": "handle homematic temperatur rightly",
    "transform": [
        {
            "operation": "shift",
            "spec": {
                "Datetime": "ts",
                "Temperature": "val",
                "Timestamp": "ts"
            }
        },
        {
            "operation": "default",
            "spec": {
                "TempUnit": "°C"
            }
        },
        {
            "operation": "timestamp",
            "spec": {
                "Datetime": {
                    "inputFormat": "$unixext",
                    "outputFormat": "2006-01-02T15:04:05-0700"
                }
            }
        }
    ]
}
```

#### Testen einer Transformationsregel eines Backends 

**POST**: https://127.0.0.1:9443/api/v1/admin/backends/{backendname}/rules/{rulename}/test

**Beschreibung**: Testen einer Transformationsregel eines Backends.

**Request**: https://127.0.0.1:9443/api/v1/admin/backends/sensors/rules/hm_temp_simple/test

Payload:

```json
{
    "val": 22.8,
    "ts": 1588142598973,
    "lc": 1588142598973
}
```

**Response**: 

```json
{
    "Timestamp": 1588142598973,
    "Datetime": "2020-04-29T08:43:18+0200",
    "Temperature": 22.8,
    "TempUnit": "°C"
}
```

### Files API

Security: Ja


#### Upload einer Datei

**Request**: **POST**: https://127.0.0.1:9443/api/v1/files/{backendname}/

**Beschreibung**: Upload einer Datei auf den Server. Dateien dürfen nicht größer sein als 10MB.

**Security role**: edit

**Request**: **POST**: https://127.0.0.1:9443/api/v1/files/sensors/

​	**Payload**: Http Formbased File Upload. Name des Formfeldes: file

**Response**:

```JSON
{
    "fileid": "5ea92df4d015d95201f6b4b8",
    "filename": "readme.md"
}
```

​	**Headers**: `Location: /api/v1/files/sensors/5ea92df4d015d95201f6b4b8`

#### Download einer Datei

**Request**: **GET**: https://127.0.0.1:9443/api/v1/files/{backendname}/{fileid}

**Beschreibung**: Download einer Datei vom Server.

**Security role**: read

**Request**: **GET**: https://127.0.0.1:9443/api/v1/files/sensors/5ea92df4d015d95201f6b4b8

**Response**:

Die Datei als Download.



