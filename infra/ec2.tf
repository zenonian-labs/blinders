resource "aws_instance" "blinders" {
  ami                    = "ami-02a2af70a66af6dfb"
  instance_type          = "t2.micro"
  key_name               = aws_key_pair.blinders_tf_ec2_key.key_name
  vpc_security_group_ids = [aws_security_group.blinders_ec2_security_group.id]

  tags = {
    Name = "blinders-server"
  }
}


resource "aws_instance" "blinders-services" {
  ami                    = "ami-02a2af70a66af6dfb"
  instance_type          = "t2.medium"
  key_name               = aws_key_pair.blinders_tf_ec2_key.key_name
  vpc_security_group_ids = [aws_security_group.blinders_ec2_security_group.id]

  root_block_device {
    volume_size = 30 # in GB
  }

  tags = {
    Name = "blinders-services"
  }
}

# TODO: need to resolve security group
resource "aws_security_group" "blinders_ec2_security_group" {
  name = "blinders-ec2-security-group"

  # Accept all inbound requests
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "all"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Accept all outbound requests
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "all"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Create RSA key of size 4096 bits
resource "tls_private_key" "blinders_tf_ec2_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "tf_ec2_key" {
  content  = tls_private_key.blinders_tf_ec2_key.private_key_pem
  filename = "${path.module}/tf_ec2_key.pem"
}

resource "aws_key_pair" "blinders_tf_ec2_key" {
  key_name   = "blinders_tf_ec2_key"
  public_key = tls_private_key.blinders_tf_ec2_key.public_key_openssh
}

output "instance_user_data" {
  description = "IP of the EC2 instance"
  value       = aws_instance.blinders.user_data
}

output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.blinders.public_ip
}

output "enable_key_file" {
  value = "chmod 400 ./tf_ec2_key.pem"
}

output "ssh_command_to_blinders_database" {
  value = "ssh ec2-user@${aws_instance.blinders.public_ip} -i ./tf_ec2_key.pem"
}


output "ssh_command_to_blinders_services" {
  value = "ssh ec2-user@${aws_instance.blinders-services.public_ip} -i ./tf_ec2_key.pem"
}
