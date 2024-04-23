# important:
# - must use runtime "provided.al2" for go lambdas (not provided.al2023)
# - handler must be "bootstrap" for runtime "provided.al2"
# - must use arc "arm64" for go lambdas

resource "aws_lambda_function" "dictionary" {
  function_name    = "${var.project.name}-dictionary-${var.project.environment}"
  filename         = "../../dist/dictionary-${var.project.environment}.zip"
  handler          = "blinders.dictionary_aws_lambda_function.lambda_handler"
  source_code_hash = filebase64sha256("../../dist/dictionary-${var.project.environment}.zip")
  role             = aws_iam_role.lambda_role.arn
  runtime          = "python3.10"
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "pysuggest" {
  function_name    = "${var.project.name}-pysuggest-${var.project.environment}"
  filename         = "../../dist/suggest-${var.project.environment}.zip"
  timeout          = 60
  handler          = "blinders.suggest_lambda.lambda_handler"
  source_code_hash = filebase64sha256("../../dist/suggest-${var.project.environment}.zip")
  role             = aws_iam_role.lambda_role.arn
  runtime          = "python3.10"
  architectures    = ["arm64"]
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]

  environment {
    variables = merge(local.envs,{
      COLLECTING_FUNCTION_NAME : aws_lambda_function.collecting.function_name
    })
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "translate" {
  function_name    = "${var.project.name}-translate-${var.project.environment}"
  filename         = "../../dist/translate-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  source_code_hash = filebase64sha256("../../dist/translate-${var.project.environment}.zip")

  environment {
    variables = merge(local.envs,{
      COLLECTING_FUNCTION_NAME : aws_lambda_function.collecting.function_name
    })
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "ws_connect" {
  function_name    = "${var.project.name}-ws-connect-${var.project.environment}"
  filename         = "../../dist/connect-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  source_code_hash = filebase64sha256("../../dist/connect-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "ws_authorizer" {
  function_name    = "${var.project.name}-ws-authorizer-${var.project.environment}"
  filename         = "../../dist/ws_authorizer-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  source_code_hash = filebase64sha256("../../dist/ws_authorizer-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "ws_disconnect" {
  function_name    = "${var.project.name}-ws-disconnect-${var.project.environment}"
  filename         = "../../dist/disconnect-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  source_code_hash = filebase64sha256("../../dist/disconnect-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "ws_chat" {
  function_name    = "${var.project.name}-ws-chat-${var.project.environment}"
  filename         = "../../dist/wschat-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  source_code_hash = filebase64sha256("../../dist/wschat-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "rest" {
  function_name    = "${var.project.name}-rest-api-${var.project.environment}"
  filename         = "../../dist/rest-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  source_code_hash = filebase64sha256("../../dist/rest-${var.project.environment}.zip")

  environment {
    variables = merge(local.envs, {
      NOTIFICATION_FUNCTION_NAME : aws_lambda_function.notification.function_name,
      EXPLORE_FUNCTION_NAME : aws_lambda_function.explore.function_name
    })
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "notification" {
  function_name = "${var.project.name}-notification-${var.project.environment}"
  filename      = "../../dist/notification-${var.project.environment}.zip"
  handler       = "bootstrap"
  role          = aws_iam_role.lambda_role.arn
  # temporily disable to prevent cycles
  # depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  source_code_hash = filebase64sha256("../../dist/notification-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}

resource "aws_lambda_function" "explore" {
  function_name = "${var.project.name}-explore-${var.project.environment}"
  filename      = "../../dist/explore-${var.project.environment}.zip"
  handler       = "bootstrap" # default for provided.al2
  role          = aws_iam_role.lambda_role.arn
  # temporily disable to prevent cycles
  # depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  source_code_hash = filebase64sha256("../../dist/explore-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}


resource "aws_lambda_function" "collecting" {
  function_name    = "${var.project.name}-collecting-${var.project.environment}"
  filename         = "../../dist/collecting-${var.project.environment}.zip"
  handler          = "bootstrap"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  # depends_on       = [aws_iam_role_policy_attachment.attach_iam_policy_to_iam_role]
  source_code_hash = filebase64sha256("../../dist/collecting-${var.project.environment}.zip")

  environment {
    variables = local.envs
  }

  tags = {
    project     = var.project.name
    environment = var.project.environment
  }
}
