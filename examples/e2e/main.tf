module "main" {
  source = "./module"
  count  = 10
  name   = "e2e-${count.index}"
}
