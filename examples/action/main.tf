# a credential of type NONE holds the whole webhook URL
resource "ilert_credential" "discord" {
  name         = "discord ops"
  type         = "NONE"
  host_url     = var.discord_webhook_url
  category     = "global-action"
  sub_category = "discord"
}

resource "ilert_action" "discord" {
  name           = "post high priority alerts to discord"
  template_key   = "discord_alert"
  entity_kind    = "ALERT"
  credential_id  = ilert_credential.discord.id
  trigger_events = ["alert.created", "alert.accepted", "alert.resolved"]
  condition_icl  = "entity.source.id in [\"${var.alert_source_id}\"] and entity.priority == \"HIGH\""

  config = jsonencode({
    pushComments = true
    editMessage  = false
  })
}
