### CoreDNS for mapping DNS Record like using in On-Premise Platform

operator-sdk init --domain quay.io --repo github.com/wdrdres3qew5ts21/coredns-integration-operator

### layout Project
Manifest Metadata ถูก Generate มาจาก `./config/manifests/bases` ถ้าอยากจะแก้อะไรให้ไปแก้ในนี้แล้วตอนรันคำสั่ง bundle build มันจะนำ metadata ไปด้วย


Bundle
https://github.com/operator-framework/operator-registry/blob/master/docs/design/operator-bundle.md



ทำการติดตั้ง Operator life Cycle ไปยัง Kbuernetes Clsuter เรา
operator-sdk olm install

Build ตัวอย่างโครงโปรเจค
```
make docker-build docker-push IMG="quay.io/linxianer12/coredns-integration-operator:0.0.4"
```

สร้าง Template Manifest กับอัพเดท Image Push ที่อยู่ใน `bundle/manifests`

```
make bundle IMG="quay.io/linxianer12/coredns-integration-operator:0.0.4"

make bundle-build bundle-push BUNDLE_IMG="quay.io/linxianer12/coredns-integration-bundle:0.0.4"
```
https://sdk.operatorframework.io/docs/overview/project-layout/


ลง Operator Catalog ใหม่พร้อมติดตั้งใน Namespace Kubernetes Context ที่เรากำลังอยู่
```
operator-sdk run bundle quay.io/linxianer12/coredns-integration-bundle:0.0.4
```
ลบการติดตั้ง Catalog 
operator-sdk cleanup coredns-integration-operator

### Check Operator Life Cycle
สำหรับ Openshift จะอยู่ที่ namespace `openshift-operator-lifecycle-manager`

```
operator-sdk olm status --olm-namespace openshift-operator-lifecycle-manager

```