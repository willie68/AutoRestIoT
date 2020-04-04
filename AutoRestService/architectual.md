# Auto Rest IoT Service #

## Anlegen eines Backends mittels yaml Datei

Dieser kleine Service ermöglicht es, schnell eine permanente Datenspeicherung über ein einfaches REST/gRPC Interface zu ermöglichen. Dazu definiert man ein sog. Backend. Jedes Backend hat einen eindeutigen Namen und definiert einen eigenen API Key. Dieser Key muss bei jeder REST Kommunikation als Header X-mcs-apikey mit gesendet werden.

Der Backendname darf max. 60 Zeichen lang sein und sollte nur aus Kleinbuchstaben bestehen. Sonderzeichen wie "$", ".", "_", oder "-" sind nicht erlaubt. Auch ein Leerzeichen " " darf nicht verwendet werden. 

Jedes Backend besteht nun aus einer Reihe von Modellen. Ein Modell kann man sich als eine Tabelle vorstellen. Will man Daten in eine Tabelle ablegen, muss man ein Modell dafür definieren.
Jedes Modell hat einen eigenen Namen und definiert eine Reihe von Feldern/Attributen. Grundsätzlich werden alle übergebenden Attribute gespeichert, auch wenn Sie hier nicht definiert wurden. Die Definition dient einerseits der besseren Indexierung. D.h. will man einen Suchindex für ein Attribute oder eine Kombination mehrerer Attribute anlegen, müssen die verwendeten Attribute hier zumindest mit Type definiert werden. 
Auch eine Attributvalidierung (wie z.B. auch das Mandatory) erfordert die Definition des jeweiligen Attributes hier.

Typische JSON Attribute/Objektverschachtelungen sind grundsätzlich erlaubt. Neben den Attributen kann man pro Modell auch noch eine Reihe von Suchindizies definieren, um einen schnelleren Zugriff zu ermöglichen. Eine Besonderheit stellt der Volltextindex dar. Man kann pro Modell **einen** Volltextindex definieren. Dabei wird dann jedes angegebene Feld in diesem Index gespeichert und über eine eingängige Suchsyntax wieder findbar abgelegt. Dazu mehr im Kapitel Suche.

Ein Service kann mehrere Backends verwalten. 

Bei der Mongo Storage Implementierung werden die verschiedenen Backends allerdings in einer Datenbank abgelegt. Einzelne Modelle werden in jeweils einer Collection abgelegt.  Der Collectionname besteht aus dem Backendnamen  "." und dem Modellnamen.

```yaml
applicationname: schematicworld  #name without whitespaces and special charaters
description: Willies World Schematics Database #description of the backend
models:  #definition of the different models
    - name: schematics #name of the models, no whitespaces or special chars
      description: This are the different schematics # a model description
      fields: #definition of the fields/attributes
        - name: manufacturer #name of the field, , no whitespaces or special chars
          type: string  #int, float, bool, map, id, more to come...
          mandantory: true #internal validator for present
          collection: false #field is a collection of types 
        - name: model
          type: string
          mandantory: true
          collection: false
        - name: tags
          type: string
          mandantory: false
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