FROM golang:1.22.0
LABEL version="1.0"
LABEL description="What's Go application"
LABEL maintainer="Dmytro Karpovych <karpovych.d.v@gmail.com>"

RUN apt-get update -qq

# You need librariy files and headers of tesseract and leptonica.
# When you miss these or LD_LIBRARY_PATH is not set to them,
# you would face an error: "tesseract/baseapi.h: No such file or directory"
RUN apt-get install -y -qq libtesseract-dev libleptonica-dev

# In case you face TESSDATA_PREFIX error, you minght need to set env vars
# to specify the directory where "tessdata" is located.
ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/

# Load languages.
# These {lang}.traineddata would b located under ${TESSDATA_PREFIX}/tessdata.
RUN apt-get install -y -qq \
  tesseract-ocr-eng \
  tesseract-ocr-ukr \
  tesseract-ocr-rus
# See https://github.com/tesseract-ocr/tessdata for the list of available languages.
# If you want to download these traineddata via `wget`, don't forget to locate
# downloaded traineddata under ${TESSDATA_PREFIX}/tessdata.

# Setup project
WORKDIR /app

# Create a user and switch to it
RUN adduser --disabled-password --gecos '' appuser
USER appuser

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN mkdir ./build

# Change ownership of the build directory to appuser
USER root
RUN chown -R appuser:appuser ./build
USER appuser

# Build the application
RUN go build -o build/whatsgo ./cmd/whatsgo

CMD ["./build/whatsgo", "--config=config/config.yaml"]
