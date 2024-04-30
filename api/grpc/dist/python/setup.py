from distutils.core import setup

setup(name='%NAME%',
      version='%VERSION%',
      description='GRPC client for %NAME%',
      author='ci',
      author_email='ci@test.com',
      packages=['%NAME%'],
      package_data={
          '%NAME%': ['*.pyi', 'py.typed'],
      },
      include_package_data=True,
      )
