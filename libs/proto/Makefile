
lint-breaking:
	buf breaking --against 'https://github.com/raghav1030/NaboServer.git#branch=main'

lint-proto:
	buf lint --config buf.yaml

generate-proto:
	buf generate --template buf.gen.yaml
