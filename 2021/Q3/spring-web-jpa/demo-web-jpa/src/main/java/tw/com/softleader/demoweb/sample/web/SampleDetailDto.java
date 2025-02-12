package tw.com.softleader.demoweb.sample.web;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Getter;
import lombok.Setter;
import lombok.ToString;

import java.time.LocalDate;

@Schema(description = "範例物件明細檔")
@Getter
@Setter
@ToString
public class SampleDetailDto {

    @Schema(description = "明細檔ID")
    Long id;

    @Schema(description = "主檔ID")
    Long sampleId;

    @Schema(description = "明細名稱")
    String name;

    @Schema(description = "明細日期")
    LocalDate date;

}
