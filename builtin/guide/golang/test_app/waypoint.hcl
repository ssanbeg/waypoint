project = "guides"

app "example" {

  build {
    use "go" {
      output_name="server"
      source="./"
    }
  }
}