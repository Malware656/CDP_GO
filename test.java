import java.io.IOException;
import java.util.Date;

public class test {
    public static void main(String[] args) throws IOException, InterruptedException {
        long start,end;
        // for(int i=0;i<10;i++){
            start = new Date().getTime();
            String command = "google-chrome --headless --disable-gpu --print-to-pdf=file1.pdf /home/calibraint/Desktop/template.html";
            Process proc = Runtime.getRuntime().exec(command);
            proc.waitFor();
            end = new Date().getTime();
            System.out.println(end-start);
        // }
    }
}