# Alert action <-> alert source relation example

This demos the `ilert_alert_action_source_attachment` virtual resource. Each instance attaches **one** alert source to a shared alert action without overwriting the other attached sources.

Use this pattern when multiple teams or modules need to attach their own alert sources to a shared alert action without fighting over the full source list.

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}"
```

To verify the non-destructive behavior: remove `team_b_to_shared` from `main.tf`, re-apply, then confirm via `curl` or the iLert UI that `team_a` is still attached to the shared alert action.
