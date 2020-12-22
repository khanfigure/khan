Static inline content
---------------------

```
# salt
/tmp/file1:
  file.managed:
    - contents: "content of file"

# chef
file '/tmp/file1' do
  content "content of file"
end

# ansible
- name: /tmp/file1
  copy:
    dest: /tmp/file1
    content: "content of file"

# puppet
file { "/tmp/file1":
  content => "content of file"
}

# khan?
file:
  path: /tmp/file1
  content: "content of file"
```

Content from ./files/
---------------------

```
# salt
/tmp/file1:
  file.managed:
    - source: salt://files/file1

# chef
cookbook_file '/tmp/file1' do
	source 'file1'
	action :create
end

# ansible
- name: /tmp/file1
  copy:
  	 dest: /tmp/file1
  	 src: files/file1

# puppet
file { "/tmp/file1":
  source => 'puppet://file1'
}

# khan?
file:
  path: /tmp/file1
  src: files/file1
```

Content from another file on host
---------------------------------

```
# salt
/tmp/file1:
  file.copy:
    - source: /opt/file1

# chef
file '/tmp/file1' do
	content IO.read('/opt/file1')
	action :create
end

# ansible
- name: /tmp/file1
  copy:
  	 dest: /tmp/file1
  	 src: /opt/file1
  	 remote_src: yes

# puppet
file { "/tmp/file1":
  source => '/opt/file1'
}

# khan?
file:
  path: /tmp/file1
  local: /opt/file1
```

Templates
---------
```yaml
template:
  path: /tmp/file1
  content: "This is a template. You are on host {{ hostname }}."

template:
  path: /tmp/file1
  src: templates/file1

template:
  path: /tmp/file1
  file: /opt/templates/file1
```

Shortcuts
---------
```yaml
file /tmp/file1:
	src: files/file1

/tmp/file1:
	src: files/file1

/tmp/file1: files/file1

template /tmp/file1: templates/file1
```
