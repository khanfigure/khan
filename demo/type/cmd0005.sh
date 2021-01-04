cat > somefiles.yaml <<EOF
file:
  path: /tmp/file.txt
  content: |
    This is the contents of a managed file.

  user: bozo
  group: bozo

user:
  name: bozo
  uid: 3000

group:
  name: bozo
  gid: 3000

EOF
