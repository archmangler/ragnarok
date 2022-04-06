//S3 bucket, policies and EC2 instance for upload access

resource "aws_s3_bucket" "source" {
  bucket        = "eqnx-tradedata-source001"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "source_access" {
  bucket              = aws_s3_bucket.source.id
  block_public_acls   = true
  block_public_policy = true
  ignore_public_acls  = true
}

resource "aws_iam_policy" "source_policy" {
  name        = "source_policy_dataops"
  path        = "/"
  description = "Allow "

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "VisualEditor0",
        "Effect" : "Allow",
        "Action" : [
          "s3:PutObject",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:DeleteObject"
        ],
        "Resource" : [
          "arn:aws:s3:::*/*",
          "arn:aws:s3:::eqnx-tradedata-source001"
        ]
      }
    ]
  })
}


//EC2 - S3 Access policy Configuration

#IAM policy to allow access from EC2 instances (later EKS service)
resource "aws_iam_role" "source_role" {
  name = "source_role_dataops"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role" "source_role_eks" {
  name = "source_role_eks_dataops"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "eks.amazonaws.com"
        }
      },
    ]
  })
}

#Attach policy to role
resource "aws_iam_role_policy_attachment" "source_bucket_policy" {
  role       = aws_iam_role.source_role.name
  policy_arn = aws_iam_policy.source_policy.arn
}

#profile for asn ec2 instance
resource "aws_iam_instance_profile" "source_profile" {
  name = "source_profile_dataops"
  role = aws_iam_role.source_role.name
}

//EC2 Data Loading / staging instance
resource "aws_instance" "data_ops_staging" {
  subnet_id                   = module.vpc.public_subnets[1]
  associate_public_ip_address = true
  ami                         = "ami-0f74c08b8b5effa56" //"ami-03abf47a104b960b4"
  instance_type               = "t2.medium"

  key_name = "data_ops_key_staging"

  //  security_groups = [ aws_security_group.worker_group_mgmt_one.id]
  security_groups = [aws_security_group.all_worker_mgmt.id, aws_security_group.worker_group_mgmt_two.id, aws_security_group.worker_group_mgmt_one.id]

  iam_instance_profile = aws_iam_instance_profile.source_profile.id

  tags = {
    Name     = "eqnx-data-staging"
    Function = "jira-eep-21233"
  }
}

resource "aws_eip" "data_ops" {
  instance = aws_instance.data_ops_staging.id
  vpc      = true
}

resource "aws_key_pair" "data_ops" {
  key_name   = "data_ops_key_staging"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDACyG2S7IJ1acCqWzCX3fYSN5aMZ6YzQKDbkhypYeQHUbe6hgJpzG9oiOHLgO6n7c33Km7Rh5Q+rzg7deyO5OWsKh1pXLrryeyIJXpQV9+sJnxYhxpAFh7bremQya1UIryAQ5NJGdxQDWKmX2QsuroAEa8RK2PiVIvB6EOfA6twfd/HvO5Ez0w2APiD+N5VS3AUjVJ0myaTYOFdVtauJj/Ix3Kzdo67dZ5K2iyK+d/J5L1IXJghxp98/UYGtIBpoFOFVgd/u+MoSln5Ey6JpMkp73vPu7iqR2gcJKJxxhEzH2dwYcnGmOhP7evaERBy7wJQ7nF7nbBi7SPtpnJyXAwrprLfUzKePIXUzBR3kMEZ1oIZoL3bSDJrFFbrc/if24rS1KYtGr2w2v3Issedc5Wwi9ERl8GVp7BEYVJQQOjAmXfEDafDC6MlAyZDbetE1MeEehC1w4R8Yjldxj8Ur1wfY0hHy9MpTo6CzvgskqP4w9YYOvN547HoyC0meF1dT04w9UMiA9on+fD+SatbQLc5TluhDd9xX9cOXG2RUTzNuZyG3ghXkcey0p7eyzRkY+xszeSN4d9TJRaxY4Dc1LyUgHhkky8XQTrLFhOkG8hDjYzlRAlQn6o8v0Toqaj2cNbySm4H0j4BNuC9j/uAc+mvYlIr1WVaH8sh/XKmeHoYw== traiano.welcome.contractor@accenture.com"
}


//Accessing S3 (from EC2 instance with correct policy):
//List      : aws s3 ls s3://eqnx-tradedata-source --recursive --human-readable --summarize
//Copy up   : aws s3 cp dmesg.log s3://eqnx-tradedata-source
//Copy down : aws s3 cp s3://eqnx-tradedata-source/dmesg.log new.log
//Copy down bulk: aws s3 cp s3://eqnx-tradedata-source/inputs ./source --recursive
//Make Dir: aws s3 cp --recursive $local_folder s3://eqnx-tradedata-source/$DATE/
//Delete    : na

