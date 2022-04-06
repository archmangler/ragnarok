variable "worker_group_mgmt_one_ingress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_block  = string
    description = string
  }))
  default = [
    {
      from_port   = 0 
      to_port     = 65434
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "elb requirement"
    },
    {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "core requirement"
    },
    {
      from_port   = 10250
      to_port     = 10250
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "core requirement"
    },
  ]
}

variable "worker_group_mgmt_two_ingress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_block  = string
    description = string
  }))
  default = [
    {
      from_port   = 0
      to_port     = 65534
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "elb requirement"
    },
    {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "core requirement"
    },
    {
      from_port   = 10250
      to_port     = 10250
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "core requirement"
    },
  ]
}

variable "all_worker_mgmt_ingress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_block  = string
    description = string
  }))
  default = [
    {
      from_port   = 0
      to_port     = 65534
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "elb requirement"
    },
    {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "core requirement"
    },
    {
      from_port   = 10250
      to_port     = 10250
      protocol    = "tcp"
      cidr_block  = "192.0.0.0/8"
      description = "core requirement"
    },
  ]
}

variable "worker_group_mgmt_one_egress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_block  = string
    description = string
  }))
  default = [
    {
      from_port   = 0
      to_port     = 65434
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "elb requirement"
    },
    {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "core requirement"
    },
    {
      from_port   = 10250
      to_port     = 10250
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "core requirement"
    },
  ]
}

variable "worker_group_mgmt_two_egress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_block  = string
    description = string
  }))
  default = [
    {
      from_port   = 0
      to_port     = 65534
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "elb requirement"
    },
    { 
      from_port   = 10250
      to_port     = 10250
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "core requirement"
    },
  ]
}

variable "all_worker_mgmt_egress_rules" {
  type = list(object({
    from_port   = number
    to_port     = number
    protocol    = string
    cidr_block  = string
    description = string
  }))
  default = [
    {
      from_port   = 0
      to_port     = 65534
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "elb requirement"
    },
    { 
      from_port   = 10250
      to_port     = 10250
      protocol    = "tcp"
      cidr_block  = "0.0.0.0/0"
      description = "core requirement"
    },
  ]
}


resource "aws_security_group_rule" "worker_group_mgmt_one_ingress_rules" {
  count             = length(var.worker_group_mgmt_one_ingress_rules)
  type              = "ingress"
  from_port         = var.worker_group_mgmt_one_ingress_rules[count.index].from_port
  to_port           = var.worker_group_mgmt_one_ingress_rules[count.index].to_port
  protocol          = var.worker_group_mgmt_one_ingress_rules[count.index].protocol
  cidr_blocks       = [var.worker_group_mgmt_one_ingress_rules[count.index].cidr_block]
  description       = var.worker_group_mgmt_one_ingress_rules[count.index].description
  security_group_id = aws_security_group.worker_group_mgmt_one.id
}

resource "aws_security_group_rule" "worker_group_mgmt_two_ingress_rules" {
  count             = length(var.worker_group_mgmt_two_ingress_rules)
  type              = "ingress"
  from_port         = var.worker_group_mgmt_two_ingress_rules[count.index].from_port
  to_port           = var.worker_group_mgmt_two_ingress_rules[count.index].to_port
  protocol          = var.worker_group_mgmt_two_ingress_rules[count.index].protocol
  cidr_blocks       = [var.worker_group_mgmt_two_ingress_rules[count.index].cidr_block]
  description       = var.worker_group_mgmt_two_ingress_rules[count.index].description
  security_group_id = aws_security_group.worker_group_mgmt_two.id
}

resource "aws_security_group_rule" "all_worker_mgmt_ingress_rules" {
  count             = length(var.all_worker_mgmt_ingress_rules)
  type              = "ingress"
  from_port         = var.all_worker_mgmt_ingress_rules[count.index].from_port
  to_port           = var.all_worker_mgmt_ingress_rules[count.index].to_port
  protocol          = var.all_worker_mgmt_ingress_rules[count.index].protocol
  cidr_blocks       = [var.all_worker_mgmt_ingress_rules[count.index].cidr_block]
  description       = var.all_worker_mgmt_ingress_rules[count.index].description
  security_group_id = aws_security_group.all_worker_mgmt.id
}

resource "aws_security_group_rule" "worker_group_mgmt_one_egress_rules" {
  count             = length(var.worker_group_mgmt_one_egress_rules)
  type              = "egress"
  from_port         = var.worker_group_mgmt_one_egress_rules[count.index].from_port
  to_port           = var.worker_group_mgmt_one_egress_rules[count.index].to_port
  protocol          = var.worker_group_mgmt_one_egress_rules[count.index].protocol
  cidr_blocks       = [var.worker_group_mgmt_one_egress_rules[count.index].cidr_block]
  description       = var.worker_group_mgmt_one_egress_rules[count.index].description
  security_group_id = aws_security_group.worker_group_mgmt_one.id
}

resource "aws_security_group_rule" "worker_group_mgmt_two_egress_rules" {
  count             = length(var.worker_group_mgmt_two_egress_rules)
  type              = "egress"
  from_port         = var.worker_group_mgmt_two_egress_rules[count.index].from_port
  to_port           = var.worker_group_mgmt_two_egress_rules[count.index].to_port
  protocol          = var.worker_group_mgmt_two_egress_rules[count.index].protocol
  cidr_blocks       = [var.worker_group_mgmt_two_egress_rules[count.index].cidr_block]
  description       = var.worker_group_mgmt_two_egress_rules[count.index].description
  security_group_id = aws_security_group.worker_group_mgmt_two.id
}

resource "aws_security_group_rule" "all_worker_mgmt_egress_rules" {
  count             = length(var.all_worker_mgmt_egress_rules)
  type              = "egress"
  from_port         = var.all_worker_mgmt_egress_rules[count.index].from_port
  to_port           = var.all_worker_mgmt_egress_rules[count.index].to_port
  protocol          = var.all_worker_mgmt_egress_rules[count.index].protocol
  cidr_blocks       = [var.all_worker_mgmt_egress_rules[count.index].cidr_block]
  description       = var.all_worker_mgmt_egress_rules[count.index].description
  security_group_id = aws_security_group.all_worker_mgmt.id
}


resource "aws_security_group" "worker_group_mgmt_one" {
  name_prefix = "worker_group_mgmt_one"
  vpc_id      = module.vpc.vpc_id

  /*
  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"

    cidr_blocks = [
      "192.0.0.0/8",
    ]
  }
*/

}

resource "aws_security_group" "worker_group_mgmt_two" {
  name_prefix = "worker_group_mgmt_two"
  vpc_id      = module.vpc.vpc_id

  /*
  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"

    cidr_blocks = [
      "192.0.0.0/16",
    ]
  }
*/

}

resource "aws_security_group" "all_worker_mgmt" {
  name_prefix = "all_worker_management"
  vpc_id      = module.vpc.vpc_id

  /*
  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"

    cidr_blocks = [
      "10.0.0.0/8",
      "172.16.0.0/12",
      "192.0.0.0/8",
    ]
  }
*/

}

//If you have EFS storage somewhere
/*
resource "aws_security_group" "efs" {
   name = "${local.cluster_name}-efs-sg"
   description= "Allow inbound efs traffic from ec2"
   vpc_id = module.vpc.vpc_id

   ingress {
     security_groups = [aws_security_group.all_worker_mgmt.id]
     from_port = 2049
     to_port = 2049 
     protocol = "tcp"
   }     
        
   egress {
     security_groups = [aws_security_group.all_worker_mgmt.id]
     from_port = 0
     to_port = 0
     protocol = "-1"
   }
 }
*/
