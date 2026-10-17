# IAM managed policy — allows invocation of the specified foundation models
# and read-only listing of available foundation models.
resource "aws_iam_policy" "bedrock_access" {
  name        = "${var.instance_name}-bedrock-access"
  description = "CSB sandbox Bedrock access policy — expires ${local.ttl_expires_at}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "BedrockInvokeModel"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream",
        ]
        Resource = local.model_arns
      },
      {
        Sid    = "BedrockListModels"
        Effect = "Allow"
        Action = [
          "bedrock:ListFoundationModels",
          "bedrock:GetFoundationModel",
        ]
        Resource = "*"
      },
      {
        Sid    = "BedrockApplyGuardrail"
        Effect = "Allow"
        Action = [
          "bedrock:ApplyGuardrail",
        ]
        Resource = aws_bedrock_guardrail.content_filter.guardrail_arn
      },
    ]
  })
}

# Content-filtering guardrail — blocks harmful categories and anonymises PII.
resource "aws_bedrock_guardrail" "content_filter" {
  name                      = "${var.instance_name}-guardrail"
  description               = "Content filtering guardrail for CSB sandbox — expires ${local.ttl_expires_at}"
  blocked_input_messaging   = "This request was blocked by the content policy."
  blocked_outputs_messaging = "This response was blocked by the content policy."

  content_policy_config {
    filters_config {
      type            = "SEXUAL"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "VIOLENCE"
      input_strength  = "MEDIUM"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "HATE"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "INSULTS"
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
    }
    filters_config {
      type            = "MISCONDUCT"
      input_strength  = "MEDIUM"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "PROMPT_ATTACK"
      input_strength  = "HIGH"
      output_strength = "NONE"
    }
  }

  sensitive_information_policy_config {
    pii_entities_config {
      type   = "EMAIL"
      action = "ANONYMIZE"
    }
    pii_entities_config {
      type   = "PHONE"
      action = "ANONYMIZE"
    }
    pii_entities_config {
      type   = "US_SOCIAL_SECURITY_NUMBER"
      action = "BLOCK"
    }
  }
}
