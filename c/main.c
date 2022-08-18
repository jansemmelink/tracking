#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>

#define M_BLOCK_SIZE                  41

typedef struct m_pos_s {
    float lat_f;
    float lon_f;
} m_pos_t;

typedef struct m_vehicle_s {
    unsigned int                id_ud;
    const char*                 reg_ac[10];
    m_pos_t                     pos_z;
    unsigned long long          ts_llud;
} m_vehicle_t;

int main() {    
    const char* fn = "../data/VehiclePositions.dat";
    FILE* f = fopen (fn, "rb");
    if (f == NULL) {
        exit (1);
    }

    //get file size
    size_t                      l_file_size_ud;
    {
        struct stat                 l_stat_z;
        stat (fn, &l_stat_z);
        l_file_size_ud = l_stat_z.st_size;
    }/*scope*/

    size_t                      l_file_ofs_ud = 0;

    unsigned char               l_buf_auc[M_BLOCK_SIZE * 2];    //*2 for left over plus new block
    size_t                      l_buf_len_ud = 0;

    int                         l_count_d = 0;
    while (l_file_ofs_ud < l_file_size_ud) {
        if (l_file_ofs_ud < l_file_size_ud - M_BLOCK_SIZE) {
            //read whole block
            size_t l_nr_blocks_read_ud = fread (
                l_buf_auc + l_buf_len_ud,
                M_BLOCK_SIZE,
                1,
                f);
            if (l_nr_blocks_read_ud != 1) {
                printf ("ERROR: ofs=%zu: failed to read block: got total=%d rem bytes=%zu\n", l_file_ofs_ud, l_count_d, l_buf_len_ud);
                exit (1);
            }/*if failed to read*/
            l_buf_len_ud += (size_t)M_BLOCK_SIZE;
        }/*if reading full block*/
        else
        {
            //end of file - read the last few bytes
            size_t l_nr_bytes_read_ud = fread (
                l_buf_auc + l_buf_len_ud,
                1,
                l_file_size_ud - l_file_ofs_ud,
                f);
            if (l_nr_bytes_read_ud == 0) {
                printf ("ERROR: ofs=%zu: failed to read last %zu bytes, got %zu\n", l_file_ofs_ud, l_file_size_ud - l_file_ofs_ud, l_nr_bytes_read_ud);
                exit (1);
            }/*if failed to read*/

            l_buf_len_ud += l_nr_bytes_read_ud;
        }

        //printf("ofs:%zu read:%d, buf_len:%zu\n", l_file_ofs_ud, M_BLOCK_SIZE, l_buf_len_ud);

        //process all complete records from buffer
        const unsigned char*        l_next_puc  = l_buf_auc;
        const unsigned char*        l_start_puc = l_next_puc;
        const unsigned char*        l_end_puc   = l_buf_auc + l_buf_len_ud;
        while (l_next_puc + 21 < l_end_puc) {
            const unsigned int*         l_id_pud = (const unsigned int*)l_next_puc;
            l_next_puc += 4;

            const char*                 l_reg_pc = (const char*)(l_next_puc);
            int incomplete = 0;
            while ((!incomplete) && (*l_next_puc != 0)) {
                l_next_puc ++;
                if (l_next_puc == l_end_puc) {
                    incomplete = 1;
                    //printf ("incomplete string/...\n");
                    break;
                }
            }
            if ((incomplete) || (l_next_puc + 16 >= l_end_puc)) {
                //printf ("incomplete lat/lon/...\n");
                break;
            }

            size_t                      l_reg_len_ud = l_next_puc - (unsigned char*)l_reg_pc;
            l_next_puc ++;  //skip '\0'

            const float*                l_lat_pf = (const float*)l_next_puc;
            l_next_puc += 4;

            const float*                l_lon_pf = (const float*)l_next_puc;
            l_next_puc += 4;
            
            const unsigned long long*   l_ts_pulld = (const unsigned long long*)l_next_puc;
            l_next_puc += 8;

            // printf("%10zu: id=%10u=0x%08x, reg(len:%zu):\"%s\", %8.2f, %8.2f, %llu %zu\n",
            //         l_file_ofs_ud + (l_start_puc - l_buf_auc),
            //         *l_id_pud, *l_id_pud,
            //         l_reg_len_ud,
            //         l_reg_pc,
            //         *l_lat_pf,
            //         *l_lon_pf,
            //         *l_ts_pulld,
            //         l_buf_len_ud);

            if (  (*l_lat_pf < -90.0)
               || (*l_lat_pf > 90.0)
               || (*l_lon_pf < -180.0)
               || (*l_lon_pf > 180.0)
               || (l_reg_len_ud != 9))
            {
                printf ("ERROR AT ofs: %zu + %zu = %zu\n",
                    l_file_ofs_ud,
                    l_start_puc - l_buf_auc,
                    l_file_ofs_ud + (l_start_puc - l_buf_auc));
                exit (1);
            }

            //todo: here we can construct the vehicle struct
            //(only could have created a byte aligned struct and read it directly too...)
            //anyway, the grid and loopup not written in C...
            //but the loading from disk with block reads is very fast

            l_count_d ++;
            l_start_puc = l_next_puc;   //only move once we used a full record
        }//while processing buffer

        size_t l_used_ud = (l_start_puc - l_buf_auc);
        if (l_used_ud % 30 != 0) {
            printf ("used=%zu ERROR\n", l_used_ud);
            exit (1);
        }

        //shift remaining data to front of buffer
        l_buf_len_ud -= l_used_ud;
        memmove (l_buf_auc, l_start_puc, l_buf_len_ud);

        //front of buffer now reflects this offset
        l_file_ofs_ud += l_used_ud;
    }//while reading
    fclose(f);
    f = NULL;

    printf("loaded %d entries\n", l_count_d);
}