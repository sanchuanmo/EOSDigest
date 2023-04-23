#include<openssl/sha.h>
#include<string.h>
#include<cmath>
#include<iostream>
#include <iomanip>
#include <sstream>



using namespace std;

class sha256{
    public:
        sha256();
        explicit sha256( const string& hex_str);

        string str()const;
        operator string()const;

        const char* data()const;
        char* data();
        size_t data_size() const {return 256/8;}

        static sha256 hash( const char* d, uint32_t dlen);

        static sha256 hash(const unsigned char* d, uint32_t dlen);

        static sha256 hash(const string& s);

        static sha256 hash(const sha256& s);

        static sha256 hash(const std::pair<sha256,sha256>& value);

        class encoder
    {
        public:
            encoder();
            ~encoder();

            void write(const char*d, uint32_t dlen);
            void reset();
            sha256 result();

        private:
            struct impl{
                SHA256_CTX ctx;
            }my;
    };
    uint64_t _hash[4];
};

