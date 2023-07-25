# <authority_id>/<provisioner_id>/<name>
AUTHORITY_ID=ed2e4f38-fd2d-4eb0-9280-52b697636873
PROVISIONER_ID=57b8ade4-5873-4a15-911c-a4fff5999600

terraform import smallstep_provisioner_webhook.devices ${AUTHORITY_ID}/${PROVISIONER_ID}/devices
