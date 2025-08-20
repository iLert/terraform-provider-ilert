resource "ilert_call_flow" "example" {
  name     = "example"
  language = "en"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ANSWERED"
      target {
        name = "Create alert"
        node_type = "CREATE_ALERT"
        metadata {
          alert_source_id = -1 // your alert source id
        }
      }
    }
  }
}

resource "ilert_call_flow" "always_on_support" {
  name = "Always-On Support"
  language = "en"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ANSWERED"
      target {
        node_type = "AUDIO_MESSAGE"
        metadata {
          text_message =  "Thank you for calling <company name>. Our support team is available 24/7 to assist you. Please hold while we connect you to the next available responder."
          ai_voice_model = "emma"
          language = "en"
        }
        branches {
          branch_type = "CATCH_ALL"
          target {
            node_type = "ROUTE_CALL"
            metadata {
              targets {
                target = "-1" // your user id
                type = "USER"
              }
              call_style = "ORDERED"
              call_timeout_sec = 45
            }
            branches {
              branch_type = "BRANCH"
              condition = "context.connectedTarget == '-1'"
              target {
                node_type = "VOICEMAIL"
                metadata {
                  text_message = "We're sorry, but all responders are currently unavailable. Please leave your name, contact information, and a brief message, and we'll get back to you as soon as possible."
                  ai_voice_model = "emma"
                  language = "en"
                }
                branches {
                  branch_type = "BRANCH"
                  condition = "context.recordedMessageUrl != null"
                  target {
                    node_type = "CREATE_ALERT"
                    metadata {
                      alert_source_id = -1 // your alert source id
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}

resource "ilert_call_flow" "business_hours_support" {
  name = "Business Hours Support"
  language = "en"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ANSWERED"
      target {
        node_type = "AUDIO_MESSAGE"
        metadata {
          text_message = "Thank you for calling <company name>."
          ai_voice_model = "emma"
          language = "en"
        }
        branches {
          branch_type = "CATCH_ALL"
          target {
            name = "Business hours"
            node_type = "SUPPORT_HOURS"
            metadata {
              support_hours_id = -1 // your support hours id
            }
            branches {
              branch_type = "BRANCH"
              condition = "context.supportHoursState == 'DURING'"
              target {
                node_type = "ROUTE_CALL"
                metadata {
                  targets {
                    target = "-1" // your user id
                    type = "USER"
                  }
                  call_style = "ORDERED"
                  call_timeout_sec = 45
                }
              }
            }
            branches {
              branch_type = "BRANCH"
              condition = "context.supportHoursState == 'OUTSIDE'"
              target {
                node_type = "VOICEMAIL"
                metadata {
                  text_message = 	"You've reached us outside of our business hours. Please leave your name, contact information, and a brief message, and we'll get back to you during our next business day. Thank you!"
                  ai_voice_model = "emma"
                  language = "en"
                }
                branches {
                  branch_type = "BRANCH"
                  condition =  "context.recordedMessageUrl != null"
                  target {
                    node_type = "CREATE_ALERT"
                    metadata {
                      alert_source_id = -1 // your alert source id
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}

resource "ilert_call_flow" "interactive_menu" {
  name = "Interactive Menuu"
  language = "en"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ANSWERED"
      target {
        node_type = "AUDIO_MESSAGE"
        metadata {
          text_message = "Thank you for calling <company name>."
          ai_voice_model = "emma"
          language = "en"
        }
        branches {
          branch_type = "CATCH_ALL"
          target {
            node_type = "IVR_MENU"
            metadata {
              text_message = "To report an incident, press 1. For system monitoring assistance, press 2."
              ai_voice_model = "emma"
              enabled_options = ["1", "2"]
              language = "en"
              retries = 1
            }
            branches {
              branch_type = "BRANCH"
              condition = "context.ivrChoice == '1'"
              target {
                node_type = "ROUTE_CALL"
                metadata {
                  targets {
                    target = "-1" // your user id
                    type = "USER"
                  }
                  call_style = "ORDERED"
                  call_timeout_sec = 45
                }
              }
            }
            branches {
              branch_type = "BRANCH"
              condition = "context.ivrChoice == '2'"
              target {
                node_type = "ROUTE_CALL"
                metadata {
                  targets {
                    target = "-1" // your user id
                    type = "USER"
                  }
                  call_style = "ORDERED"
                  call_timeout_sec = 45
                }
              }
            }
          }
        }
      }
    }
  }
}