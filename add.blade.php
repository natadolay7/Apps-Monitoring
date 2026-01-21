@extends('new2024.layout.master')
@section('back')
    {{ url('input-lesson-to-learn') }}
@endsection
@section('title')
    Input Lesson To Learn  
@endsection
@section('style')
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/dropzone/5.9.3/dropzone.min.css" />
@endsection
@section('button_back')
    <li class="breadcrumb-item"><a href="{{ url('input-lesson-to-learn') }}">Lesson To Learn</a></li>
@endsection
@section('content')
    <style>
        #preview img {
            height: 100px;
            margin: 10px;
            border-radius: 5px;
            object-fit: cover;
            border: 1px solid #ccc;
        }
    </style>

    <section class="content">
        <div class="container-fluid">
            <div class="row">
                <div class="col-md-8">
                    <div class="card card-outline card-info">
                        <div class="card-body">
                            <form id="uploadForm" method="POST" action="{{ url('input-lesson-to-learn/store') }}"
                                enctype="multipart/form-data">
                                {{ csrf_field() }}
                                <div class="row">
                                    <div class="col-md-8 mb-3">
                                        <label for="">Cover</label>
                                        <input type="file" name="cover" class="form-control">
                                    </div>
                                    <div class="col-md-8 mb-3">
                                        <label for="">File PDF</label>
                                        <input type="file" class="form-control" name="file_pdf" accept=".pdf">
                                    </div>
                                    <div class="col-md-8 mb-3">
                                        <label for="">Title</label>
                                        <input type="text" name="judul" class="form-control" placeholder="Judul" required>

                                    </div>
                                    <div class="col-md-8 mb-3">
                                        <label for="">Wording Singkat</label>
                                        <input type="text" name="wording_singkat" class="form-control" placeholder="Judul" required>

                                    </div>

                                    <div class="col-md-12">
                                        <label for="">Deskripsi</label>
                                        <textarea id="summernote2" name="deskripsi" required>   
                                        </textarea>
                                    </div>

                                    <div class="col-md-12 mb-3">
                                        <label>Slider (Klik atau Drag & Drop)</label>
                                        <!-- Drag area -->
                                        <div id="my-dropzone" class="dropzone dz-clickable">
                                            <div class="dz-message">Drag & drop file di sini atau klik untuk pilih</div>
                                        </div>
                                        <!-- Input hidden untuk nyimpan file drag-and-drop -->
                                        <input type="file" id="real-upload" name="slider[]" multiple style="display:none;"
                                            required>
                                    </div>

                                    <div class="col-md-12 mt-4">
                                        <button class="btn btn-primary">Simpan</button>
                                    </div>

                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
@endsection
@section('script')
    <script src="https://cdnjs.cloudflare.com/ajax/libs/dropzone/5.9.3/dropzone.min.js"></script>
    <script>
        Dropzone.autoDiscover = false;

        const realInput = document.getElementById('real-upload');

        const myDropzone = new Dropzone("#my-dropzone", {
            url: "#",
            autoProcessQueue: false,
            clickable: true,
            maxFiles: 15,
            addRemoveLinks: true,
            acceptedFiles: "image/*",
            init: function() {
                this.on("addedfile", function(file) {
                    // Limit 15 file
                    if (this.files.length > 15) {
                        this.removeFile(file);
                        alert("Maksimal hanya 15 file.");
                        return;
                    }

                    // Update input[type=file]
                    const dataTransfer = new DataTransfer();
                    this.files.forEach(f => dataTransfer.items.add(f));
                    realInput.files = dataTransfer.files;

                    // Tambahkan event click untuk preview
                    const img = file.previewElement.querySelector("img");
                    if (img) {
                        img.style.cursor = "pointer";
                        img.addEventListener("click", function() {
                            const reader = new FileReader();
                            reader.onload = function(e) {
                                window.open(e.target.result, "_blank");
                            };
                            reader.readAsDataURL(file);
                        });
                    }
                });

                this.on("removedfile", function(file) {
                    const dataTransfer = new DataTransfer();
                    this.files.forEach(f => dataTransfer.items.add(f));
                    realInput.files = dataTransfer.files;
                });
            }
        });
    </script>



    <script>
        $(function() {
            $('#summernote2').summernote({
                toolbar: [
                    ['style', ['bold', 'italic', 'underline', 'clear']],
                    ['font', ['strikethrough', 'superscript', 'subscript']],
                    ['fontsize', ['fontsize']],
                    ['color', ['color']],
                    ['para', ['ul', 'ol', 'paragraph']],
                    ['height', ['height']]
                    // Tidak ada ['insert', ['picture']] ← tombol upload gambar dihapus
                ]
            });
        });
    </script>
@endsection
