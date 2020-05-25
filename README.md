# go_wss_mtls
Example of mutually authenticated TLS/SSL TCP connection Client/Server for IOT device service control.

Dependencies on gorilla websockets:
<code> go get github.com/gorilla/websocket </code>
<code> go get github.com/gorilla/mux </code>

Run echo server:
<code> go run -v src/server/server.go </code>

Run client:
<code> go run -v src/client/client.go </code>


![PlantUML model](http://www.plantuml.com/plantuml/png/dP71IiGm48RlUOgXdgMhnq1bhO8eYorKyBXCsxWDJPkIJDS5yTrjjyt67FMOcV_xyfDiBQ6XiLFCjh8Vq607EHgIbHomfnphMlO7kDtysgVwvEOt6yPAVO8epI8sU0vIh5hHruKmHrc9OFKkIZkDXM5J02PwLJndR_0evdW4LpjvQ2XLOuW-A2bw2aQvGlhGflCobYG9F0c2-oDAXQKIAsSXILXO3AxlV1yE3T70urZf2bMZBV516uhHUVRLuc4NPEP3uqNoKcIlVBqPTj8IBCqAGaEmfl_PwFnfUxiGyPUiluDAxbOEdzRoU8cxhrXsSyez7OTpWBcpDt38MW_uNawSjNzuYYcydPPZftu0)