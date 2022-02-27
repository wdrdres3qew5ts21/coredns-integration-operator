### CoreDNS for mapping DNS Record like using in On-Premise Platform

operator-sdk init --domain quay.io --repo github.com/wdrdres3qew5ts21/coredns-integration-operator

### layout Project
Bundle
https://github.com/operator-framework/operator-registry/blob/master/docs/design/operator-bundle.md


ทำการติดตั้ง Operator life Cycle ไปยัง Kbuernetes Clsuter เรา
operator-sdk olm install

Build ตัวอย่างโครงโปรเจค
```
make docker-build docker-push IMG="quay.io/linxianer12/coredns-integration-operator:0.0.2"
```

```
สร้าง Template Manifest
make bundle IMG="quay.io/linxianer12/coredns-integration-operator:0.0.2"
make bundle-build bundle-push BUNDLE_IMG="quay.io/linxianer12/coredns-integration-bundle:0.0.2"
```
https://sdk.operatorframework.io/docs/overview/project-layout/


ลง Operator Catalog ใหม่
operator-sdk run bundle quay.io/linxianer12/coredns-integration-bundle:0.0.2

ลบการติดตั้ง Catalog 
operator-sdk cleanup coredns-integration-operator

### Check Operator Life Cycle
สำหรับ Openshift จะอยู่ที่ namespace `openshift-operator-lifecycle-manager`

```
operator-sdk olm status --olm-namespace openshift-operator-lifecycle-manager

```