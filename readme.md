### CoreDNS for mapping DNS Record like using in On-Premise Platform

Version 0.0.54 เป็น Version ที่สามารถ Reconcile Service กับ Deployment ได้แล้วและแก้บั้คหา Resource ไม่เจอได้สำเร็จด้วยการให้ selector Client สามารถทำแบบ Dynamic ได้ตาม Resource ที่เป็นคนสร้างขึ้นมา

1. ใช้คำสั่ง `make install` เพื่อติดตั้ง CRD File Manifest ทั้งหมดไปยัง Cluster โดยไม่ต้อง Deploy Push ไปที่ Image Registry แต่วิธีนี้จะไม่ถูกจัดการผ่าน Operator Life Cycle Management และไม่เห็น UI ใน Openshift
2. ใช้คำสั่ง `make run` เพื่อ run manager controller pod local ที่ laptop เราเพื่อ debug Operator โดยไม่ต้อง Deploy Application จริงๆผ่าน Operator Life Cycle Manager


### Create API and Resource

https://www.brighttalk.com/webcast/18106/470697?utm_source=brighttalk-portal&utm_medium=web&utm_campaign=channel-feed

https://www.youtube.com/watch?v=89PdRvRUcPU&t=798s


Base Image: https://quay.io/repository/openshift/origin-coredns?tab=info

สร้าง Project ครั้งแรกด้วยคำสั่ง
```
operator-sdk init --domain quay.io --repo github.com/wdrdres3qew5ts21/coredns-integration-operator
```
Reload CoreDNS แบบทันทีด้วยการ Kill Pod
https://docs.microsoft.com/en-us/azure/aks/coredns-custom

## layout Project

### Manifest Metadata Config
ถูก Generate มาจาก `./config/manifests/bases` ถ้าอยากจะแก้อะไรให้ไปแก้ในนี้แล้วตอนรันคำสั่ง bundle build มันจะนำ metadata ไปด้วย

### Operator Controller Coding
อยู่ใน directory `./controllers`


Bundle
https://github.com/operator-framework/operator-registry/blob/master/docs/design/operator-bundle.md


ทำการติดตั้ง Operator life Cycle ไปยัง Kubernetes Cluster เราถ้าเกิดใช้ Vanila Kubernetes แล้วไม่มี Operator
```
operator-sdk olm install
```

Build ตัวอย่างโครงโปรเจค
```
make docker-build docker-push IMG="quay.io/linxianer12/coredns-integration-operator:0.0.6"
```

เพิ่ม API Custom Resource

operator-sdk create api --group cache --version v1alpha1 --kind DNS --resource --controller


สร้าง Template Manifest กับอัพเดท Image Push ที่อยู่ใน `bundle/manifests`

```
make bundle IMG="quay.io/linxianer12/coredns-integration-operator:0.0.6"

make bundle-build bundle-push BUNDLE_IMG="quay.io/linxianer12/coredns-integration-bundle:0.0.6"
```
https://sdk.operatorframework.io/docs/overview/project-layout/


เพิ่มสิทธิ์ให้กับ Golang Operator ถ้าเจอเรื่อง `is forbidden: cannot set blockOwnerDeletion if an ownerReference refers to a resource you can’t set finalizers on:`
https://sdk.operatorframework.io/docs/faqs/

โดยให้เพิ่ม RBAC Marker ให้ตรงกับ Resource ของเราและตรวจสอบ Noun ให้ดีๆด้วยว่ามีเติม S เป็นแบบ Plural หรือเปล่าเพราะถ้าไม่มีแล้วเราไปเติม S มันก็จะผิดนั่นเอง
https://book.kubebuilder.io/reference/markers/rbac.html
```
+kubebuilder:rbac:groups=cache.quay.io,resources=dns/finalizers,verbs=update
```

### ลง Operator Catalog ใหม่พร้อมติดตั้งใน Namespace Kubernetes Context ที่เรากำลังอยู่
ใช้คำสั่งเดียวจบ
```
export IMAGE_VERSION=<Current Version>
operator-sdk run bundle quay.io/linxianer12/coredns-integration-bundle:$IMAGE_VERSION
```
ลบการติดตั้ง Catalog 
```
operator-sdk cleanup coredns-integration-operator
```
### Check Operator Life Cycle
สำหรับ Openshift จะอยู่ที่ namespace `openshift-operator-lifecycle-manager`

```
operator-sdk olm status --olm-namespace openshift-operator-lifecycle-manager

```
### Setup Go Environment
https://gist.github.com/vsouza/77e6b20520d07652ed7d


### Build Script
เราจะทดสอบใน `private-dns` Namespace
```
oc new-project private-dns
oc apply -f oc apply -f permission 

operator-sdk cleanup coredns-integration-operator

export IMAGE_VERSION=0.0.54

./build-push-operator.sh 

operator-sdk cleanup coredns-integration-operator

operator-sdk run bundle quay.io/linxianer12/coredns-integration-bundle:$IMAGE_VERSION
```

reference API Version 
https://github.com/deepak1725/hello-operator2

Create ConfigMap Setup Manager
https://techbloc.net/archives/4630