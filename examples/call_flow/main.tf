resource "ilert_call_flow" "example" {
  name     = "example-call-flow"
  language = "English"

  root_node = [{
    node_type = "ROOT"
    name      = "root"

    branches = [
      {
        branch_type = "ANSWERED"
        target = [{
          node_type = "AUDIO_MESSAGE"
          name      = "welcome"
          metadata = [{
            text_message = "Welcome to our hotline."
          }]
        }]
      },
      {
        branch_type = "CATCH_ALL"
        target = [{
          node_type = "VOICEMAIL"
          name      = "vm"
          metadata = [{
            text_message = "Please leave a message after the beep."
          }]
        }]
      }
    ]
  }]
}
