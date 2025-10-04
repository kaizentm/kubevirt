qemu-img create -f raw -o size=10G /tmp/cidata.raw
kubectl cp  ubuntu-24.04-minimal-cloudimg-amd64.img  virt-launcher-testvm2-j6zcq:/tmp/ 
kubectl cp  work/kubevirt/test.xml  virt-launcher-testvm2-tlctd:/tmp/

qemu-system-x86_64 \
  -machine q35,accel=ACCEL \
  -cpu host -m 1024 \
  -drive if=virtio,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -device virtio-serial-pci \
  -chardev stdio,id=con0,signal=off,server=off \
  -device virtconsole,chardev=con0 \
  -nographic -serial none

qemu-system-x86_64 \
  -machine q35,accel=tcg \
  -cpu host -m 1024 \
  -drive if=virtio,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -device virtio-serial-pci \
  -chardev stdio,id=con0,signal=off \
  -device virtconsole,chardev=con0 \
  -nographic -serial none -monitor none  


sudo qemu-system-x86_64 \
  -machine q35,accel=tcg \
  -cpu qemu64 \
  -m 1024 \
  -drive if=virtio,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -device virtio-serial-pci \
  -chardev stdio,id=con0,signal=off \
  -device virtconsole,chardev=con0 \
  -nographic -serial none -monitor none  


qemu-system-x86_64 \
  -machine q35,accel=tcg \
  -cpu qemu64 -m 1024 \
  -drive if=virtio,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -serial stdio \
  -nographic -monitor none 



# working
qemu-system-x86_64 \
  -machine q35,accel=tcg \
  -cpu qemu64 -m 1024 \
  -drive if=virtio,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -serial stdio \
  -device virtio-serial-pci \
  -chardev pty,id=vc0 \
  -device virtconsole,chardev=vc0 \
  -nographic -monitor none  

# not working
qemu-system-x86_64 \
  -machine q35,accel=mshv \
  -cpu host -m 1024 \
  -drive if=virtio,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -serial stdio \
  -device virtio-serial-pci \
  -chardev pty,id=vc0 \
  -device virtconsole,chardev=vc0 \
  -nographic -monitor none    

# working
qemu-system-x86_64 \
  -machine pc,accel=tcg \
  -cpu qemu64 -m 512 -smp 1 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -serial stdio -display none -no-reboot

# not working  
qemu-system-x86_64 \
  -machine pc,accel=mshv \
  -cpu qemu64 -m 512 -smp 1 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -serial stdio -display none -no-reboot


sudo qemu-system-x86_64 \
  -machine pc,accel=mshv \
  -cpu qemu64 \
  -m 512 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -serial stdio \
  -display none \
  -monitor none \
  -no-reboot  

############################################################
# Additional mshv debugging variants (run one at a time)   #
# Aim: isolate region map + interrupt vector 0 warnings.   #
############################################################

# Variant A: Add tracing + isa-debugcon to capture very early output
sudo qemu-system-x86_64 \
  -machine pc,accel=mshv \
  -cpu qemu64 \
  -m 512 \
  -smp 1 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -chardev file,id=dbg,path=/tmp/mshv-debugcon.log,append=on \
  -device isa-debugcon,iobase=0x402,chardev=dbg \
  -serial stdio \
  -d guest_errors,int,ioport \
  -D /tmp/qemu-mshv-min.log \
  -display none -monitor none -no-reboot

# Variant B: Force userspace irqchip (kernel-irqchip=off)
sudo qemu-system-x86_64 \
  -machine pc,accel=mshv,kernel-irqchip=off \
  -cpu qemu64 \
  -m 512 -smp 1 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -serial stdio -display none -monitor none -no-reboot

# Variant C: Split irqchip (if off still fails)
sudo qemu-system-x86_64 \
  -machine pc,accel=mshv,kernel-irqchip=split \
  -cpu qemu64 \
  -m 512 -smp 1 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -serial stdio -display none -monitor none -no-reboot

# Variant D: Remove most defaults to find conflicting device mapping
# NOTE: With -nodefaults we must add an IDE controller + disk explicitly.
sudo qemu-system-x86_64 \
  -nodefaults \
  -machine pc,accel=mshv \
  -cpu qemu64 -m 512 -smp 1 \
  -device piix3-ide \
  -drive id=drive0,if=none,file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,cache=none \
  -device ide-hd,drive=drive0,bus=ide.0 \
  -serial stdio \
  -chardev file,id=dbg,path=/tmp/mshv-nodefaults-debugcon.log,append=on \
  -device isa-debugcon,iobase=0x402,chardev=dbg \
  -d guest_errors,int,ioport \
  -D /tmp/qemu-mshv-nodefaults.log \
  -display none -monitor none -no-reboot

# Variant E: Same as A but host CPU (to see if host model aggravates issue)
sudo qemu-system-x86_64 \
  -machine pc,accel=mshv \
  -cpu host \
  -m 512 -smp 1 \
  -drive file=/home/azureuser/work/kubevirt/cirros-0.6.3-x86_64-disk.img,format=qcow2,if=ide,cache=none \
  -chardev file,id=dbg,path=/tmp/mshv-hostcpu-debugcon.log,append=on \
  -device isa-debugcon,iobase=0x402,chardev=dbg \
  -serial stdio \
  -d guest_errors,int,ioport \
  -D /tmp/qemu-mshv-hostcpu.log \
  -display none -monitor none -no-reboot

# After running a variant, collect:
#  1) tail -50 /tmp/mshv-*.log (qemu trace)
#  2) hexdump -C /tmp/mshv*-debugcon.log | head (early BIOS/kernel chars?)
#  3) Elapsed seconds until you conclude 'hang'.
# Post first variant that still reproduces the warnings so we can pivot.