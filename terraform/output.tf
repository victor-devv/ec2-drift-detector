output "instance_id" {
  value = aws_instance.sample_instance.id
}

output "instance_private_ip" {
  value = aws_instance.sample_instance.private_ip
}
