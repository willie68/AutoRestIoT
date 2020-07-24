module github.com/willie68/AutoRestIoT

go 1.14

require (
	github.com/aphistic/golf v0.0.0-20180712155816-02c07f170c5a
	github.com/aphistic/sweet v0.3.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/docgen v1.0.5
	github.com/go-chi/render v1.0.1
	github.com/go-delve/delve v1.4.0 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/goiiot/libmqtt v0.9.5 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/hashicorp/consul v1.7.2 // indirect
	github.com/hashicorp/consul/api v1.5.0
	github.com/joyent/triton-go v0.0.0-20180628001255-830d2b111e62
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/qntfy/jsonparser v1.0.2 // indirect
	github.com/qntfy/kazaam v3.4.8+incompatible
	github.com/rs/zerolog v1.18.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/willie68/kazaam v3.4.8+incompatible
	go.mongodb.org/mongo-driver v1.3.4
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	gopkg.in/square/go-jose.v2 v2.5.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)


replace github.com/qntfy/kazaam => ../kazaam