variable "instance_name" { type = string }
variable "resource_tags_json" { type = string }
variable "region" { type = string }
variable "user_name" { type = string }
variable "policy_arn" { type = string }
variable "guardrail_id" { type = string }
variable "guardrail_arn" { type = string }
variable "available_models" { type = string }
variable "bedrock_endpoint" { type = string }
variable "ttl_expires_at" { type = string }
variable "cf_context_json" { type = string }
variable "cf_originating_identity_json" { type = string }
variable "cf_app_guid" {
	type    = string
	default = ""
}
