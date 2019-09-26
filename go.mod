module git.stuhome.com/Sunmxt/wing

go 1.12

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20181218151757-9b75e4fe745a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
)

require (
	github.com/RichardKnop/machinery v1.6.9
	github.com/Sunmxt/form v1.5.1
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/gin-contrib/sessions v0.0.0-20190512062852-3cb4c4f2d615
	github.com/gin-gonic/gin v1.4.0
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jinzhu/configor v1.0.0
	github.com/jinzhu/gorm v1.9.8
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d
	github.com/rakyll/statik v0.1.6
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/text v0.3.2
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/bluesuncorp/validator.v5 v5.10.3
	gopkg.in/bluesuncorp/validator.v8 v8.18.2
	gopkg.in/bluesuncorp/validator.v9 v9.29.0
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/ldap.v3 v3.0.3
	gopkg.in/src-d/go-git.v4 v4.13.1
	k8s.io/api v0.0.0-20190515023547-db5a9d1c40eb
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.1 // indirect
	k8s.io/utils v0.0.0-20190520173318-324c5df7d3f0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
