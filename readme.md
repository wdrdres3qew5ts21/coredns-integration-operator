operator-sdk init --domain quay.io --repo github.com/wdrdres3qew5ts21/coredns-integration-operator


ทำการติดตั้ง Operator life Cycle ไปยัง Kbuernetes Clsuter เรา
operator-sdk olm install

Build ตัวอย่างโครงโปรเจค
```
make docker-build docker-push IMG="quay.io/linxianer12/coredns-integration-operator:0.0.1"
```

```
สร้าง Template Manifest
make bundle IMG="quay.io/linxianer12/coredns-integration-operator:0.0.1"
make bundle-build bundle-push BUNDLE_IMG="quay.io/linxianer12/coredns-integration-bundle:v0.0.1"
```
https://sdk.operatorframework.io/docs/overview/project-layout/

CoreDNS for mapping DNS Record like using in On-Premise Platform