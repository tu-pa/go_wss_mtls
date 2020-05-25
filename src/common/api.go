package common

// Control Channel Request/Response //
const (
	Reqid_start_service          = "start_service"
	Reqid_stop_service           = "stop_service"
	Reqid_get_avail_clients      = "get_available_clients"
	Reqid_get_client_statistics  = "get_client_statistics"
	Reqid_get_client_ip_addr     = "get_client_ip_addr"
	Reqid_start_client_data_test = "start_data_test"
	Reqid_stop_client_data_test  = "stop_get_data_test"
)

type Request struct {
	ReqId string
}

type Response struct {
	ReqId  string
	Result string
}
