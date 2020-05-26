# go_wss_mtls
Example of mutually authenticated TLS/SSL TCP connection Client/Server for IOT device service control.

Dependencies on gorilla websockets:
<code> go get github.com/gorilla/websocket </code>
<code> go get github.com/gorilla/mux </code>

Run echo server:
<code> go run -v src/server/server.go </code>

Run client:
<code> go run -v src/client/client.go </code>


![PlantUML model](http://www.plantuml.com/plantuml/png/dPBDYjim4CVlUefXU-tYxHuDBXid4592RUf2Zc4Ygx68B1bfd2QKldlbsh9moQLxrCp_usV9RuxGKjygvYvITsYnXT9F99STQYgnhgK-m-jBt4DkRc7-wOMnaxd1KruymOUzF3UqjNXdNOo07Fb5wBeIzYgvMAmEukJyM5Zc1U23fhHTyHqsOf27r5prI-jQIQ5fCIeLqWzZsnYMPTaez7gZjU04hyy7lCEgfmQoZ6b30em7Y2WV9qSwMoh1Uok68rcZODsWdois1Jz_ZjuKVZgN9abZ7AMTiPJmCUDHKGghkWOoWR0qHet8Mq6mkg9KU59YMhi1TtcL_rGtH9tlLeQZYW0OSevyp7cCyarWlG2PTqFBidUk-b8LNzFWscrWBnt1-0aLeIK8eEz3TqF6qOsEv9UiVo-eGd6uz97chsbgYKyq_nkJW8Lpp4cXk3n-qnW_QFxZPhpupI_xLlq1)