# Configuración de CloudWatch para la aplicación serverless

# Grupo de logs para la función Lambda
resource "aws_cloudwatch_log_group" "lambda_log_group" {
  name              = "/aws/lambda/uala-microblog-${var.environment}"
  retention_in_days = 14
  tags = {
    Environment = var.environment
    Application = "uala-microblog"
  }
}

# Métricas personalizadas para monitorear el rendimiento
resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  alarm_name          = "uala-microblog-lambda-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "60"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "Este alarma se activa cuando hay errores en la función Lambda"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    FunctionName = "uala-microblog-${var.environment}"
  }
}

# Alarma para latencia alta
resource "aws_cloudwatch_metric_alarm" "lambda_duration" {
  alarm_name          = "uala-microblog-lambda-duration-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = "60"
  statistic           = "Average"
  threshold           = "1000"
  alarm_description   = "Este alarma se activa cuando la duración promedio de la función Lambda supera los 1000ms"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    FunctionName = "uala-microblog-${var.environment}"
  }
}

# Alarma para throttling
resource "aws_cloudwatch_metric_alarm" "lambda_throttles" {
  alarm_name          = "uala-microblog-lambda-throttles-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "Throttles"
  namespace           = "AWS/Lambda"
  period              = "60"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "Este alarma se activa cuando hay throttling en la función Lambda"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    FunctionName = "uala-microblog-${var.environment}"
  }
}

# Alarma para API Gateway 4xx errores
resource "aws_cloudwatch_metric_alarm" "api_4xx_errors" {
  alarm_name          = "uala-microblog-api-4xx-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "4XXError"
  namespace           = "AWS/ApiGateway"
  period              = "60"
  statistic           = "Sum"
  threshold           = "10"
  alarm_description   = "Este alarma se activa cuando hay más de 10 errores 4XX en un minuto"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    ApiName = "uala-microblog-api-${var.environment}"
    Stage   = var.environment
  }
}

# Alarma para API Gateway 5xx errores
resource "aws_cloudwatch_metric_alarm" "api_5xx_errors" {
  alarm_name          = "uala-microblog-api-5xx-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "5XXError"
  namespace           = "AWS/ApiGateway"
  period              = "60"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "Este alarma se activa cuando hay errores 5XX"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  dimensions = {
    ApiName = "uala-microblog-api-${var.environment}"
    Stage   = var.environment
  }
}

# Dashboard para visualizar el estado de la aplicación
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "uala-microblog-${var.environment}"

  dashboard_body = <<EOF
{
  "widgets": [
    {
      "type": "metric",
      "x": 0,
      "y": 0,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/Lambda", "Invocations", "FunctionName", "uala-microblog-${var.environment}" ],
          [ ".", "Errors", ".", "." ],
          [ ".", "Throttles", ".", "." ],
          [ ".", "Duration", ".", ".", { "stat": "Average" } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "Lambda Metrics",
        "period": 60
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 6,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/ApiGateway", "Count", "ApiName", "uala-microblog-api-${var.environment}", "Stage", "${var.environment}" ],
          [ ".", "4XXError", ".", ".", ".", "." ],
          [ ".", "5XXError", ".", ".", ".", "." ],
          [ ".", "Latency", ".", ".", ".", ".", { "stat": "Average" } ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "API Gateway Metrics",
        "period": 60
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 12,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/DynamoDB", "ConsumedReadCapacityUnits", "TableName", "uala-microblog-users-${var.environment}" ],
          [ ".", "ConsumedWriteCapacityUnits", ".", "." ],
          [ ".", "ConsumedReadCapacityUnits", "TableName", "uala-microblog-tweets-${var.environment}" ],
          [ ".", "ConsumedWriteCapacityUnits", ".", "." ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "DynamoDB Metrics",
        "period": 60
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 18,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "AWS/ElastiCache", "CPUUtilization", "CacheClusterId", "uala-microblog-cache-${var.environment}" ],
          [ ".", "NetworkBytesIn", ".", "." ],
          [ ".", "NetworkBytesOut", ".", "." ],
          [ ".", "CurrConnections", ".", "." ]
        ],
        "view": "timeSeries",
        "stacked": false,
        "region": "${var.aws_region}",
        "title": "ElastiCache Metrics",
        "period": 60
      }
    }
  ]
}
EOF
}

# Tópico SNS para alertas
resource "aws_sns_topic" "alerts" {
  name = "uala-microblog-alerts-${var.environment}"
}

# Suscripción de email al tópico SNS
resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}