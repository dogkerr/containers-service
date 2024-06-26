// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "lintang"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/containers": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Mendapatkan semua swarm service milik user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "Mendapatkan semua swarm service milik user",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.getUserContainersResp"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "User Membuat swarm service lewat endpoint ini",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "User Membuat swarm service lewat endpoint inieperti pada postman (bearer access token saja",
                "parameters": [
                    {
                        "description": "request body membuat container",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/router.createServiceReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.createContainerResp"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/create/schedule": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "menjadwalkan pembuatan container",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "menjadwalkan pembuatan container",
                "parameters": [
                    {
                        "description": "request body penjadwalan pembuatan container",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/router.scheduleCreateReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.deleteRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/upload": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "User Membuat swarm service tetapi source code (tarfile) nya dia upload  lewat endpoint ini",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "User Membuat swarm service tetapi source code (tarfile) nya dia upload ,lewat endpoint inieperti pada postman (bearer access token saja",
                "parameters": [
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "collectionFormat": "csv",
                        "name": "env",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "imageName",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "replica",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.createContainerResp"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/{id}": {
            "get": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Mendapatkan swarm service user berdasarkan id",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "Mendapatkan swarm service user berdasarkan id",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.getContainerRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "update swarm service user (bisa juga vertical scaling disini)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "update swarm service user (bisa juga vertical scaling disini)",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "request body update container",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/router.createServiceReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.updateRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "delete user swarm service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "delete user swarm service",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.deleteRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/{id}/scale": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "horizontal scaling container user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "horizontal scaling container user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "request body horizontal scaling",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/router.scaleReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.updateRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/{id}/schedule": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "menjadwalkan start/stop/terminate container",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "menjadwalkan start/stop/terminate container",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "request body penjadwalan start/stop/terminate container",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/router.scheduleContainerReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.deleteRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/{id}/start": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "run container user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "run container user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.getContainerRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        },
        "/containers/{id}/stop": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "stop container user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "containers"
                ],
                "summary": "stop container user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "container id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/router.deleteRes"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/router.ResponseError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "domain.Container": {
            "type": "object",
            "properties": {
                "all_container_lifecycles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.ContainerLifecycle"
                    }
                },
                "container_port": {
                    "type": "integer"
                },
                "created_at": {
                    "type": "string"
                },
                "endpoint": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.Endpoint"
                    }
                },
                "env": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "id": {
                    "description": "ini cuma id row di table container",
                    "type": "string"
                },
                "image": {
                    "type": "string"
                },
                "labels": {
                    "description": "/ field dibawah ini cuma dari docker engine \u0026\u0026 bukan dari db\ntapi kalau container udah diterminate gak bisa fetch field dibawah ini",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "limit": {
                    "$ref": "#/definitions/domain.Resource"
                },
                "name": {
                    "type": "string"
                },
                "public_port": {
                    "type": "integer"
                },
                "replica": {
                    "type": "integer"
                },
                "replica_available": {
                    "description": "from docker",
                    "type": "integer"
                },
                "reservation": {
                    "$ref": "#/definitions/domain.Resource"
                },
                "service_id": {
                    "description": "id dari containernya/servicenya",
                    "type": "string"
                },
                "status": {
                    "$ref": "#/definitions/domain.ServiceStatus"
                },
                "terminated_time": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "domain.ContainerAction": {
            "type": "string",
            "enum": [
                "CREATE",
                "START",
                "STOP",
                "TERMINATE"
            ],
            "x-enum-varnames": [
                "CreateContainer",
                "StartContainer",
                "StopContainer",
                "TerminateContainer"
            ]
        },
        "domain.ContainerLifecycle": {
            "type": "object",
            "properties": {
                "containerId": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "replica": {
                    "type": "integer"
                },
                "start_time": {
                    "type": "string"
                },
                "status": {
                    "$ref": "#/definitions/domain.ContainerStatus"
                },
                "stop_time": {
                    "type": "string"
                }
            }
        },
        "domain.ContainerStatus": {
            "type": "string",
            "enum": [
                "RUN",
                "STOP"
            ],
            "x-enum-varnames": [
                "ContainerStatusRUN",
                "ContainerStatusSTOPPED"
            ]
        },
        "domain.Endpoint": {
            "description": "port container",
            "type": "object",
            "properties": {
                "protocol": {
                    "type": "string",
                    "default": "tcp"
                },
                "published_port": {
                    "type": "integer"
                },
                "target_port": {
                    "type": "integer"
                }
            }
        },
        "domain.Resource": {
            "description": "ini resource cpus \u0026 memory buat setiap container nya",
            "type": "object",
            "properties": {
                "cpus": {
                    "description": "cpu dalam milicpu (1000 cpus = 1 vcpu)",
                    "type": "integer"
                },
                "memory": {
                    "description": "memory dalam satuan mb (1000mb = 1gb)",
                    "type": "integer"
                }
            }
        },
        "domain.ServiceStatus": {
            "type": "string",
            "enum": [
                "CREATED",
                "RUN",
                "STOPPED",
                "TERMINATED"
            ],
            "x-enum-varnames": [
                "ServiceCreated",
                "ServiceRun",
                "ServiceStopped",
                "ServiceTerminated"
            ]
        },
        "domain.TimeFormat": {
            "type": "string",
            "enum": [
                "MONTH",
                "DAY",
                "HOUR",
                "MINUTE",
                "SECOND"
            ],
            "x-enum-varnames": [
                "Month",
                "Day",
                "Hour",
                "Minute",
                "Second"
            ]
        },
        "router.ResponseError": {
            "description": "error message",
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "router.createContainerResp": {
            "description": "response body endpoint membuat container",
            "type": "object",
            "properties": {
                "container": {
                    "$ref": "#/definitions/domain.Container"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "router.createServiceReq": {
            "description": "request body untuk membuat container",
            "type": "object",
            "required": [
                "endpoint",
                "image",
                "limit",
                "name",
                "replica"
            ],
            "properties": {
                "endpoint": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.Endpoint"
                    }
                },
                "env": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "image": {
                    "type": "string"
                },
                "labels": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "limit": {
                    "$ref": "#/definitions/domain.Resource"
                },
                "name": {
                    "type": "string"
                },
                "replica": {
                    "type": "integer"
                },
                "reservation": {
                    "$ref": "#/definitions/domain.Resource"
                }
            }
        },
        "router.deleteRes": {
            "description": "response body yg isinnya message success doang",
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "router.getContainerRes": {
            "description": "mendapatkan container user berdasarkan id container",
            "type": "object",
            "properties": {
                "container": {
                    "$ref": "#/definitions/domain.Container"
                }
            }
        },
        "router.getUserContainersResp": {
            "description": "response GetUsersContainer",
            "type": "object",
            "properties": {
                "containers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.Container"
                    }
                }
            }
        },
        "router.scaleReq": {
            "description": "request body horizontal scaling",
            "type": "object",
            "properties": {
                "replica": {
                    "type": "integer"
                }
            }
        },
        "router.scheduleContainerReq": {
            "description": "request body menjadwalkan start/stop/terminate container",
            "type": "object",
            "required": [
                "action",
                "id",
                "scheduled_time",
                "time_format"
            ],
            "properties": {
                "action": {
                    "$ref": "#/definitions/domain.ContainerAction"
                },
                "id": {
                    "type": "string"
                },
                "scheduled_time": {
                    "type": "integer"
                },
                "time_format": {
                    "$ref": "#/definitions/domain.TimeFormat"
                }
            }
        },
        "router.scheduleCreateReq": {
            "description": "request body penjadwalan pembuatan container",
            "type": "object",
            "required": [
                "action",
                "container",
                "scheduled_time",
                "time_format"
            ],
            "properties": {
                "action": {
                    "$ref": "#/definitions/domain.ContainerAction"
                },
                "container": {
                    "$ref": "#/definitions/router.scheduleCreateServiceReq"
                },
                "scheduled_time": {
                    "type": "integer"
                },
                "time_format": {
                    "$ref": "#/definitions/domain.TimeFormat"
                }
            }
        },
        "router.scheduleCreateServiceReq": {
            "type": "object",
            "properties": {
                "endpoint": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/domain.Endpoint"
                    }
                },
                "env": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "image": {
                    "type": "string"
                },
                "labels": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "limit": {
                    "$ref": "#/definitions/domain.Resource"
                },
                "name": {
                    "type": "string"
                },
                "replica": {
                    "type": "integer"
                },
                "reservation": {
                    "$ref": "#/definitions/domain.Resource"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "router.updateRes": {
            "description": "response body isinya message success doang",
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "103.175.219.0:8888",
	BasePath:         "/api/v1",
	Schemes:          []string{"http"},
	Title:            "go-container-service-lintang",
	Description:      "container service dogker",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
