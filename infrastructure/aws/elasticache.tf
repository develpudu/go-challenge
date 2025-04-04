terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  # Configure your desired region
  region = "us-east-1" # Example region, change as needed
  # Ensure your AWS credentials are configured (e.g., via environment variables, AWS config file)
}

# Fetch default VPC for simplicity - Replace with specific VPC if needed
resource "aws_default_vpc" "default" {}

# Fetch default subnets for simplicity - Replace with specific subnets if needed
resource "aws_default_subnet" "subnet_a" {
  availability_zone = "us-east-1a" # Adjust AZ as needed
}

resource "aws_default_subnet" "subnet_b" {
  availability_zone = "us-east-1b" # Adjust AZ as needed
}

# Create an ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "microblog_cache_subnet_group" {
  name       = "microblog-cache-subnet-group"
  subnet_ids = [aws_default_subnet.subnet_a.id, aws_default_subnet.subnet_b.id]

  tags = {
    Project = "Microblogging Platform"
    Purpose = "ElastiCache Subnet Group"
  }
}

# Create a Security Group for ElastiCache
# WARNING: This allows all inbound traffic on the Redis port from within the default VPC.
# Restrict this further in a production environment.
resource "aws_security_group" "microblog_cache_sg" {
  name        = "microblog-cache-sg"
  description = "Allow Redis traffic within VPC for Microblogging Platform"
  vpc_id      = aws_default_vpc.default.id

  ingress {
    from_port   = 6379 # Default Redis port
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [aws_default_vpc.default.cidr_block] # Allow from within the VPC
    # For more security, you might restrict this to the Lambda function's security group ID
    # security_groups = [aws_security_group.lambda_sg.id] # Assuming a lambda_sg exists
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Project = "Microblogging Platform"
    Purpose = "ElastiCache Security Group"
  }
}

# Create the ElastiCache Redis Replication Group (non-clustered mode)
resource "aws_elasticache_replication_group" "microblog_redis" {
  replication_group_id          = "microblog-redis-replication-group"
  description                   = "Redis cache for Microblogging Platform"
  node_type                     = "cache.t3.micro" # Choose an appropriate instance type
  number_cache_clusters         = 1                # For non-clustered mode
  engine                        = "redis"
  engine_version                = "7.0"            # Choose a Redis version
  port                          = 6379
  automatic_failover_enabled    = false            # Typically false for single-node setups
  subnet_group_name             = aws_elasticache_subnet_group.microblog_cache_subnet_group.name
  security_group_ids            = [aws_security_group.microblog_cache_sg.id]
  # parameter_group_name        = "default.redis7" # Optional: specify parameter group
  # snapshot_retention_limit    = 7                # Optional: configure backups
  # apply_immediately           = true             # Optional

  tags = {
    Project = "Microblogging Platform"
    Purpose = "Redis Cache"
  }
}

# Output the primary endpoint address for the Redis cluster
output "redis_primary_endpoint_address" {
  description = "The connection endpoint for the primary node of the Redis replication group"
  value       = aws_elasticache_replication_group.microblog_redis.primary_endpoint_address
}

output "redis_primary_endpoint_port" {
  description = "The port number of the primary node"
  value       = aws_elasticache_replication_group.microblog_redis.port
} 