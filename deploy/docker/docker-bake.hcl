variable "IAM_APP_IMAGE_NAME" {
  default = "magistral/iam:local"
}

variable "SAAS_APP_IMAGE_NAME" {
  default = "magistral/saas:local"
}

variable "RESOURCES_APP_IMAGE_NAME" {
  default = "magistral/resources:local"
}

variable "STUDENT_APP_IMAGE_NAME" {
  default = "magistral/student:local"
}

variable "LESSON_APP_IMAGE_NAME" {
  default = "magistral/lesson:local"
}

variable "FRONTEND_APP_IMAGE_NAME" {
  default = "magistral/frontend:local"
}

variable "PLANNING_APP_IMAGE_NAME" {
  default = "magistral/planning:local"
}

variable "PAYMENT_APP_IMAGE_NAME" {
  default = "magistral/payment:local"
}

variable "DOCUMENT_APP_IMAGE_NAME" {
  default = "magistral/document:local"
}

variable "REPORTING_APP_IMAGE_NAME" {
  default = "magistral/reporting:local"
}

target "iam" {
  context    = "."
  dockerfile = "deploy/docker/iam/Dockerfile"
  tags       = ["${IAM_APP_IMAGE_NAME}"]
}

target "saas" {
  context    = "."
  dockerfile = "deploy/docker/saas/Dockerfile"
  tags       = ["${SAAS_APP_IMAGE_NAME}"]
}

target "resources" {
  context    = "."
  dockerfile = "deploy/docker/resources/Dockerfile"
  tags       = ["${RESOURCES_APP_IMAGE_NAME}"]
}

target "student" {
  context    = "."
  dockerfile = "deploy/docker/student/Dockerfile"
  tags       = ["${STUDENT_APP_IMAGE_NAME}"]
}

target "lesson" {
  context    = "."
  dockerfile = "deploy/docker/lesson/Dockerfile"
  tags       = ["${LESSON_APP_IMAGE_NAME}"]
}

target "frontend" {
  context    = "."
  dockerfile = "deploy/docker/frontend/Dockerfile"
  tags       = ["${FRONTEND_APP_IMAGE_NAME}"]
}

target "planning" {
  context    = "."
  dockerfile = "deploy/docker/planning/Dockerfile"
  tags       = ["${PLANNING_APP_IMAGE_NAME}"]
}

target "payment" {
  context    = "."
  dockerfile = "deploy/docker/payment/Dockerfile"
  tags       = ["${PAYMENT_APP_IMAGE_NAME}"]
}

target "document" {
  context    = "."
  dockerfile = "deploy/docker/document/Dockerfile"
  tags       = ["${DOCUMENT_APP_IMAGE_NAME}"]
}

target "reporting" {
  context    = "."
  dockerfile = "deploy/docker/reporting/Dockerfile"
  tags       = ["${REPORTING_APP_IMAGE_NAME}"]
}

group "all" {
  targets = ["iam", "saas", "resources", "lesson", "student", "payment", "document", "reporting", "planning", "frontend"]
}
