use core::panic::PanicInfo;

#[panic_handler]
fn panic(_info: &PanicInfo -> ! {
    if let Some(location) = info.localtion(){
     println!(
           "Panicked at {}:{} {}",
           location.file(),
           location.line(),
           info.message().unwrap()
         );
    }else {
      println!("Panicked: {}",info.message().unwrap());
    }
    shutdown()
}
